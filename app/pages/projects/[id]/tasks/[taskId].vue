<script setup lang="ts">
import type { Task } from '~/types/kanban'
import { useRoute, useRouter } from 'vue-router'
import { useKanban } from '~/composables/useKanban'
import { useProjects } from '~/composables/useProjects'

const route = useRoute()
const router = useRouter()
const projectId = computed(() => String(route.params.id))
const taskId = computed(() => String(route.params.taskId))
const boardId = computed(() => String(route.query.board || 'default'))

const { board, setColumns, updateTask } = useKanban(projectId.value, boardId.value)
const { members } = useProjects()

// Locate task and column
const current = computed(() => {
  const col = board.value.columns.find(c => c.tasks.some(t => t.id === taskId.value))
  if (!col) {
    return { col: null as any, task: null as any, index: -1 }
  }
  const index = col.tasks.findIndex(t => t.id === taskId.value)
  return { col, task: col.tasks[index] as Task, index }
})

watchEffect(() => {
  if (!current.value.task) {
    // If task not found, go back to board
    router.replace(`/projects/${projectId.value}/board`)
  }
})

// Local editable state
const title = ref('')
const description = ref('')
const labels = ref<string[]>([])
const priority = ref<Task['priority']>(undefined)
const status = ref<string>('')
const assigneeIds = ref<string[]>([])
const subtasks = ref<Array<{ id: string, title: string, done: boolean }>>([])
interface CommentAttachment { id: string, name: string, size: number, type: string, isImage: boolean, url?: string }
const comments = ref<Array<{ id: string, author: { id: string, name: string, avatar?: string }, body: string, createdAt: number, attachments?: CommentAttachment[] }>>([])
const commentDraft = ref('')
const commentImages = ref<CommentAttachment[]>([])
const commentFileInput = ref<HTMLInputElement | null>(null)

function loadFromTask(t: Task) {
  title.value = t.title
  description.value = t.description || ''
  labels.value = (t.labels || []).slice()
  priority.value = t.priority
  status.value = current.value.col?.id || t.status || ''
  assigneeIds.value = (t.assignees || []).map(a => a.id)
  subtasks.value = (t.subtasks || []).map(s => ({ id: s.id, title: s.title, done: !!s.done }))
  comments.value = (t.comments || []).map(c => ({
    ...c,
    createdAt: typeof c.createdAt === 'number' ? c.createdAt : Date.now(),
    attachments: (c as any).attachments || [],
  }))
}

watchEffect(() => {
  if (current.value.task)
    loadFromTask(current.value.task)
})

// Save logic: update fields, and if status (column) changed, move task across columns preserving id/createdAt
function saveTask() {
  if (!current.value.task)
    return
  const srcCol = current.value.col
  const srcIndex = current.value.index
  const patch: Partial<Task> = {
    title: title.value.trim() || current.value.task.title,
    description: description.value.trim(),
    labels: labels.value,
    priority: priority.value,
    assignees: members.value.filter(m => assigneeIds.value.includes(m.id)),
    subtasks: subtasks.value.slice(),
    comments: comments.value.slice() as any,
    status: status.value,
  }
  const destColId = status.value || srcCol.id
  if (destColId === srcCol.id) {
    // In-place update
    updateTask(srcCol.id, current.value.task.id, patch)
  }
  else {
    const cols = [...board.value.columns]
    const from = cols.find(c => c.id === srcCol.id)
    const to = cols.find(c => c.id === destColId)
    if (!from || !to)
      return
    const task = from.tasks.splice(srcIndex, 1)[0]
    Object.assign(task, patch)
    // Ensure status updated
    task.status = destColId
    to.tasks.unshift(task)
    setColumns(cols)
  }
}

// Comments
function addComment() {
  const body = commentDraft.value.trim()
  if (!body && !commentImages.value.length)
    return
  const author = members.value[0] || { id: 'u_self', name: 'You' }
  comments.value.push({ id: `c-${Date.now()}`, author, body, createdAt: Date.now(), attachments: commentImages.value.slice() })
  commentDraft.value = ''
  commentImages.value = []
  if (commentFileInput.value)
    commentFileInput.value.value = ''
  saveTask()
}

function onCommentImagesSelected(e: Event) {
  const input = e.target as HTMLInputElement
  const files = input.files
  if (!files)
    return
  Array.from(files).forEach((file) => {
    if (!file.type.startsWith('image/'))
      return
    const att: CommentAttachment = {
      id: `catt-${Date.now()}-${Math.random().toString(36).slice(2, 7)}`,
      name: file.name,
      size: file.size,
      type: file.type,
      isImage: true,
    }
    const reader = new FileReader()
    reader.onload = () => {
      att.url = reader.result as string
      commentImages.value.push(att)
    }
    reader.readAsDataURL(file)
  })
  // allow selecting same file set again
  if (commentFileInput.value)
    commentFileInput.value.value = ''
}

function removePendingImage(id: string) {
  const i = commentImages.value.findIndex(a => a.id === id)
  if (i !== -1)
    commentImages.value.splice(i, 1)
}

// Subtasks
const subtaskOpen = ref(false)
const subtaskTitle = ref('')
const editingIndex = ref(-1)

function openNewSubtask() {
  editingIndex.value = -1
  subtaskTitle.value = ''
  subtaskOpen.value = true
}
function openEditSubtask(i: number) {
  editingIndex.value = i
  subtaskTitle.value = subtasks.value[i]?.title || ''
  subtaskOpen.value = true
}
function saveSubtask() {
  const t = subtaskTitle.value.trim()
  if (!t)
    return
  if (editingIndex.value === -1)
    subtasks.value.push({ id: `sub-${Date.now()}-${Math.random().toString(36).slice(2, 6)}`, title: t, done: false })
  else
    subtasks.value.splice(editingIndex.value, 1, { ...subtasks.value[editingIndex.value], title: t })
  subtaskOpen.value = false
  subtaskTitle.value = ''
  editingIndex.value = -1
  saveTask()
}
function toggleSubtask(i: number) {
  const s = subtasks.value[i]
  if (!s)
    return
  subtasks.value.splice(i, 1, { ...s, done: !s.done })
  saveTask()
}
function removeSubtask(i: number) {
  subtasks.value.splice(i, 1)
  saveTask()
}

// Navigation
function goBack() {
  router.push(`/projects/${projectId.value}/board`)
}
</script>

<template>
  <div class="-m-4 lg:-m-6 h-full flex flex-col">
    <!-- Header -->
    <div class="flex h-[56px] items-center gap-2 border-b px-4 justify-between">
      <div class="flex items-center gap-3 min-w-0">
        <Button size="icon" variant="ghost" title="Back" @click="goBack">
          <Icon name="i-lucide-arrow-left" />
        </Button>
        <div class="flex items-center gap-2 min-w-0">
          <Badge variant="secondary">
            {{ current.task?.id }}
          </Badge>
          <Input v-model="title" class="min-w-[240px]" placeholder="Task title" />
        </div>
      </div>
      <div class="flex items-center gap-2">
        <Select v-model="status">
          <SelectTrigger class="w-[180px]">
            <SelectValue placeholder="Status" />
          </SelectTrigger>
          <SelectContent>
            <SelectItem v-for="c in board.columns" :key="c.id" :value="c.id">
              {{ c.title }}
            </SelectItem>
          </SelectContent>
        </Select>
        <Button @click="saveTask">
          <Icon name="i-lucide-save" />
          Save
        </Button>
      </div>
    </div>

    <!-- Content -->
    <div class="grid grid-cols-1 lg:grid-cols-3 gap-6 p-4">
      <Card class="lg:col-span-2">
        <CardHeader>
          <CardTitle>Description</CardTitle>
        </CardHeader>
        <CardContent>
          <Textarea v-model="description" rows="8" placeholder="Add a detailed description..." />
        </CardContent>
      </Card>

      <Card>
        <CardHeader>
          <CardTitle>Assignees</CardTitle>
        </CardHeader>
        <CardContent class="space-y-2">
          <DropdownMenu>
            <DropdownMenuTrigger as-child>
              <Button variant="outline" class="justify-start w-full">
                <div class="flex items-center gap-2">
                  <template v-if="assigneeIds.length">
                    <div class="flex -space-x-2">
                      <Avatar v-for="a in members.filter(m => assigneeIds.includes(m.id)).slice(0, 3)" :key="a.id" class="size-6 ring-2 ring-background">
                        <AvatarImage :src="a.avatar" :alt="a.name" />
                        <AvatarFallback>{{ a.name?.slice(0, 2) }}</AvatarFallback>
                      </Avatar>
                    </div>
                    <span class="text-xs text-muted-foreground">{{ assigneeIds.length }} selected</span>
                  </template>
                  <template v-else>
                    <Icon name="i-lucide-user-plus" />
                    <span class="text-sm text-muted-foreground">Select assignees</span>
                  </template>
                </div>
              </Button>
            </DropdownMenuTrigger>
            <DropdownMenuContent class="w-56">
              <div class="px-2 py-1.5 text-xs text-muted-foreground">
                Staff
              </div>
              <DropdownMenuSeparator />
              <DropdownMenuCheckboxItem
                v-for="m in members"
                :key="m.id"
                :checked="assigneeIds.includes(m.id)"
                @update:checked="(v: boolean) => { const set = new Set(assigneeIds); v ? set.add(m.id) : set.delete(m.id); assigneeIds = Array.from(set) }"
              >
                <div class="flex items-center gap-2">
                  <Avatar class="size-5">
                    <AvatarImage :src="m.avatar" :alt="m.name" />
                    <AvatarFallback>{{ m.name?.slice(0, 2) }}</AvatarFallback>
                  </Avatar>
                  <span class="text-sm">{{ m.name }}</span>
                </div>
              </DropdownMenuCheckboxItem>
            </DropdownMenuContent>
          </DropdownMenu>
        </CardContent>
      </Card>

      <Card class="lg:col-span-2">
        <CardHeader class="flex items-center justify-between">
          <CardTitle>Subtasks</CardTitle>
          <Button size="sm" variant="outline" @click="openNewSubtask">
            <Icon name="lucide:plus" /> Add subtask
          </Button>
        </CardHeader>
        <CardContent>
          <div v-if="!subtasks.length" class="text-sm text-muted-foreground">
            No subtasks yet.
          </div>
          <ul class="space-y-2">
            <li v-for="(s, i) in subtasks" :key="s.id" class="flex items-center gap-2 group">
              <Checkbox :checked="s.done" @update:checked="() => toggleSubtask(i)" />
              <button class="text-sm flex-1 text-left truncate hover:underline" @click="openEditSubtask(i)">
                {{ s.title }}
              </button>
              <Button size="icon-sm" variant="ghost" class="opacity-0 group-hover:opacity-100" title="Delete" @click="removeSubtask(i)">
                <Icon name="lucide:trash-2" />
              </Button>
            </li>
          </ul>
        </CardContent>
      </Card>

      <Card class="lg:col-span-1">
        <CardHeader>
          <CardTitle>Comments</CardTitle>
        </CardHeader>
        <CardContent class="space-y-3">
          <div class="flex items-start gap-2">
            <Textarea v-model="commentDraft" rows="3" placeholder="Write a comment..." />
          </div>
          <div class="flex items-center gap-2">
            <input ref="commentFileInput" type="file" class="hidden" accept="image/*" multiple @change="onCommentImagesSelected">
            <Button size="sm" variant="outline" @click="commentFileInput?.click()">
              <Icon name="lucide:paperclip" /> Attach images
            </Button>
          </div>
          <div v-if="commentImages.length" class="grid grid-cols-4 gap-2">
            <div v-for="img in commentImages" :key="img.id" class="relative group border rounded overflow-hidden">
              <img v-if="img.url" :src="img.url" alt="preview" class="w-full h-20 object-cover">
              <button class="absolute top-1 right-1 bg-background/80 rounded p-0.5 opacity-0 group-hover:opacity-100 transition" title="Remove" @click="removePendingImage(img.id)">
                <Icon name="lucide:x" class="w-4 h-4" />
              </button>
            </div>
          </div>
          <div class="flex justify-end">
            <Button size="sm" :disabled="!commentDraft.trim() && !commentImages.length" @click="addComment">
              <Icon name="i-lucide-send" /> Add
            </Button>
          </div>
          <Separator />
          <div v-if="!comments.length" class="text-sm text-muted-foreground">
            No comments yet.
          </div>
          <ul class="space-y-3">
            <li v-for="c in comments" :key="c.id" class="rounded-md border p-2">
              <div class="text-xs text-muted-foreground">
                {{ new Date(c.createdAt).toLocaleString() }} — {{ c.author?.name || 'User' }}
              </div>
              <div class="text-sm whitespace-pre-wrap">
                {{ c.body }}
              </div>
              <div v-if="c.attachments?.length" class="mt-2 grid grid-cols-3 gap-2">
                <a
                  v-for="att in c.attachments"
                  :key="att.id"
                  :href="att.url"
                  target="_blank"
                  rel="noreferrer"
                  class="block border rounded overflow-hidden"
                >
                  <img v-if="att.isImage && att.url" :src="att.url" :alt="att.name" class="w-full h-24 object-cover">
                  <div v-else class="p-2 text-xs flex items-center gap-1 truncate">
                    <Icon name="lucide:file" /> {{ att.name }}
                  </div>
                </a>
              </div>
            </li>
          </ul>
        </CardContent>
      </Card>
    </div>

    <!-- Subtask Modal -->
    <Dialog v-model:open="subtaskOpen">
      <DialogContent class="sm:max-w-[420px]">
        <DialogHeader>
          <DialogTitle>{{ editingIndex === -1 ? 'Add Subtask' : 'Edit Subtask' }}</DialogTitle>
        </DialogHeader>
        <div class="space-y-3 py-2">
          <Label for="stitle">Title</Label>
          <Input id="stitle" v-model="subtaskTitle" placeholder="Subtask title" />
        </div>
        <div class="flex items-center gap-2">
          <DialogClose as-child>
            <Button variant="secondary">
              Cancel
            </Button>
          </DialogClose>
          <DialogClose as-child>
            <Button :disabled="!subtaskTitle.trim()" @click="saveSubtask">
              Save
            </Button>
          </DialogClose>
        </div>
      </DialogContent>
    </Dialog>
  </div>
</template>
