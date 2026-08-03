<script setup lang="ts">
import type { DateValue } from '@internationalized/date'
import type { UseTimeAgoMessages, UseTimeAgoOptions, UseTimeAgoUnitNamesDefault } from '@vueuse/core'
import type { Column, NewTask, Task } from '~/types/kanban'
import {
  CalendarDateTime,
  DateFormatter,
  getLocalTimeZone,
} from '@internationalized/date'
import { useRouter } from 'vue-router'
import Draggable from 'vuedraggable'
import { useKanban } from '~/composables/useKanban'
import { useProjects } from '~/composables/useProjects'
import CardFooter from '../ui/card/CardFooter.vue'

const props = defineProps<{ projectId?: string, boardId?: string }>()
const { board, addTask, updateTask, removeTask, setColumns, removeColumn, updateColumn, addColumn } = useKanban(props.projectId ?? 'default', props.boardId ?? 'default')
const { members } = useProjects()
const router = useRouter()

const df = new DateFormatter('en-US', {
  dateStyle: 'medium',
})
const dueDate = ref<DateValue | undefined>()
const dueTime = ref<string | undefined>('00:00')
const startDate = ref<DateValue | undefined>()

watch(() => dueTime.value, (newVal) => {
  if (!newVal)
    return
  if (dueDate.value) {
    const [hours, minutes] = newVal.split(':').map(Number)
    dueDate.value = new CalendarDateTime(
      dueDate.value.year,
      dueDate.value.month,
      dueDate.value.day,
      hours,
      minutes,
    )
  }
})

const showModalTask = ref<{ type: 'create' | 'edit', open: boolean, columnId: string | null, taskId?: string | null }>({
  type: 'create',
  open: false,
  columnId: null,
  taskId: null,
})
// Track the column where the task currently resides so we can move across columns on status change
const originalColumnId = ref<string | null>(null)
const newTask = reactive<NewTask>({
  title: '',
  description: '',
  priority: undefined,
  dueDate: undefined,
  status: '',
  labels: undefined,
})
type AttachmentItem = NonNullable<Task['attachments']>[number]
type CommentItem = NonNullable<Task['comments']>[number]

const assigneeIds = ref<string[]>([])
// Structured subtask items managed via a dedicated modal
const subtaskItems = ref<Array<{ id: string, title: string, done: boolean }>>([])
// Subtask modal state
const openSubtask = ref(false)
const subtaskMode = ref<'create' | 'edit'>('create')
const subtaskForm = reactive({ title: '', index: -1 })
const attachments = ref<AttachmentItem[]>([])
const commentDraft = ref('')
const comments = ref<CommentItem[]>([])
const fileInput = ref<HTMLInputElement | null>(null)
function resetData() {
  dueDate.value = undefined
  dueTime.value = '00:00'
  startDate.value = undefined
  assigneeIds.value = []
  subtaskItems.value = []
  attachments.value = []
  comments.value = []
  commentDraft.value = ''
  originalColumnId.value = null
}
watch(() => showModalTask.value.open, (newVal) => {
  if (!newVal)
    resetData()
})

function openNewTask(colId: string) {
  showModalTask.value = { type: 'create', open: true, columnId: colId }
  newTask.title = ''
  newTask.description = ''
  newTask.priority = undefined
  assigneeIds.value = []
  subtaskItems.value = []
  newTask.status = colId
  originalColumnId.value = colId
}
function createTask() {
  if (!showModalTask.value.columnId || !newTask.title.trim())
    return
  // allow selecting a different status (column) before creating
  const targetColId = (newTask.status && String(newTask.status)) || showModalTask.value.columnId
  const payload: NewTask = {
    title: newTask.title.trim(),
    description: newTask.description?.trim(),
    priority: newTask.priority,
    startDate: startDate.value?.toDate(getLocalTimeZone()),
    dueDate: dueDate.value?.toDate(getLocalTimeZone()),
    status: targetColId,
    labels: newTask.labels,
    assignees: members.value.filter(m => assigneeIds.value.includes(m.id)),
    subtasks: subtaskItems.value.map((s, i) => ({ id: s.id || `sub-${i + 1}`, title: s.title, done: !!s.done })),
    attachments: attachments.value,
    comments: comments.value,
    parentId: newTask.parentId || null,
  }
  addTask(targetColId, payload)
  showModalTask.value.open = false
}

function editTask() {
  if (!showModalTask.value.columnId || !newTask.title.trim())
    return
  const toColId = (newTask.status && String(newTask.status)) || showModalTask.value.columnId
  const fromColId = originalColumnId.value || showModalTask.value.columnId
  const patch: Partial<Task> = {
    title: newTask.title.trim(),
    description: newTask.description?.trim(),
    priority: newTask.priority,
    startDate: startDate.value?.toDate(getLocalTimeZone()),
    dueDate: dueDate.value?.toDate(getLocalTimeZone()),
    status: toColId,
    labels: newTask.labels,
    assignees: members.value.filter(m => assigneeIds.value.includes(m.id)),
    subtasks: subtaskItems.value.map((s, i) => ({ id: s.id || `sub-${i + 1}`, title: s.title, done: !!s.done })),
    attachments: attachments.value,
    comments: comments.value,
    parentId: newTask.parentId || null,
  }
  if (toColId === fromColId) {
    // same column – simple patch
    updateTask(fromColId, showModalTask.value.taskId!, patch)
  }
  else {
    // move across columns while preserving id/createdAt
    const cols = board.value.columns
    const src = cols.find(c => c.id === fromColId)
    const dst = cols.find(c => c.id === toColId)
    if (!src || !dst) {
      return
    }
    const idx = src.tasks.findIndex(t => t.id === showModalTask.value.taskId)
    if (idx === -1) {
      return
    }
    const task = src.tasks[idx]
    // remove from source
    src.tasks.splice(idx, 1)
    // apply patch
    Object.assign(task, patch)
    // insert at top of destination
    dst.tasks.unshift(task)
    // persist by setting columns
    setColumns([...cols])
  }
  showModalTask.value.open = false
}

function showEditTask(colId: string, taskId: string) {
  // Navigate to a dedicated task details page
  const pid = props.projectId ?? 'default'
  const bid = props.boardId ?? 'default'
  router.push(`/projects/${encodeURIComponent(pid)}/tasks/${encodeURIComponent(taskId)}?board=${encodeURIComponent(bid)}`)
}

// ----- Subtasks modal handlers -----
function openCreateSubtask() {
  subtaskMode.value = 'create'
  subtaskForm.title = ''
  subtaskForm.index = -1
  openSubtask.value = true
}

function openEditSubtask(index: number) {
  subtaskMode.value = 'edit'
  subtaskForm.index = index
  subtaskForm.title = subtaskItems.value[index]?.title || ''
  openSubtask.value = true
}

function saveSubtask() {
  const title = subtaskForm.title.trim()
  if (!title)
    return
  if (subtaskMode.value === 'create') {
    subtaskItems.value.push({ id: `sub-${Date.now()}-${Math.random().toString(36).slice(2, 6)}`, title, done: false })
  }
  else if (subtaskMode.value === 'edit' && subtaskForm.index > -1) {
    const curr = subtaskItems.value[subtaskForm.index]
    if (curr)
      subtaskItems.value.splice(subtaskForm.index, 1, { ...curr, title })
  }
  openSubtask.value = false
}

function removeSubtask(index: number) {
  subtaskItems.value.splice(index, 1)
}

function toggleSubtask(index: number, checked: boolean) {
  const curr = subtaskItems.value[index]
  if (!curr)
    return
  subtaskItems.value.splice(index, 1, { ...curr, done: checked })
}

function addComment() {
  const body = commentDraft.value.trim()
  if (!body)
    return
  const author = members.value[0] || { id: 'u_self', name: 'You' }
  comments.value.push({ id: `c-${Date.now()}`, author, body, createdAt: Date.now() })
  commentDraft.value = ''
}

function onFilesSelected(e: Event) {
  const input = e.target as HTMLInputElement
  const files = input.files
  if (!files)
    return
  Array.from(files).forEach((file) => {
    const isImage = file.type.startsWith('image/')
    const att: AttachmentItem = {
      id: `att-${Date.now()}-${Math.random().toString(36).slice(2, 7)}`,
      name: file.name,
      size: file.size,
      type: file.type,
      isImage,
    }
    if (isImage) {
      const reader = new FileReader()
      reader.onload = () => {
        att.url = reader.result as string
        attachments.value.push(att)
      }
      reader.readAsDataURL(file)
    }
    else {
      attachments.value.push(att)
    }
  })
  // reset input
  if (fileInput.value)
    fileInput.value.value = ''
}

function onColumnDrop(evt: any) {
  // Full columns re-ordered
  setColumns(evt.to.__draggable_component__.modelValue)
}

function renameColumn(id: string) {
  const titleRef = document.getElementById(`col-title-${id}`) as HTMLElement
  if (titleRef)
    setTimeout(() => titleRef.focus(), 500)
}

function onUpdateColumn(evt: any, id: string) {
  if (!evt.target.textContent?.trim())
    return
  updateColumn(id, evt.target.textContent?.trim())
}

function onTaskDrop() {
  // ensure state is persisted after any move (within or across columns)
  nextTick(() => setColumns([...board.value.columns]))
}

function colorPriority(p?: Task['priority']) {
  if (!p)
    return 'text-warning'
  if (p === 'low')
    return 'text-blue-500'
  if (p === 'medium')
    return 'text-warning'
  return 'text-destructive'
}

function iconPriority(p?: Task['priority']) {
  if (!p)
    return 'lucide:equal'
  if (p === 'low')
    return 'lucide:chevron-down'
  if (p === 'medium')
    return 'lucide:equal'
  return 'lucide:chevron-up'
}

const SHORT_MESSAGES = {
  justNow: 'now',
  past: (n: string, _isPast: boolean) => n,
  future: (n: string, _isPast: boolean) => n,
  invalid: '',

  second: (n: number, _isPast: boolean) => `${n}sec`,
  minute: (n: number, _isPast: boolean) => `${n}min`,
  hour: (n: number, _isPast: boolean) => `${n}h`,
  day: (n: number, _isPast: boolean) => `${n}d`,
  week: (n: number, _isPast: boolean) => `${n}w`,
  month: (n: number, _isPast: boolean) => `${n}m`,
  year: (n: number, _isPast: boolean) => `${n}y`,
} as const satisfies UseTimeAgoMessages<UseTimeAgoUnitNamesDefault>

const OPTIONS: UseTimeAgoOptions<false, UseTimeAgoUnitNamesDefault> = {
  messages: SHORT_MESSAGES,
  showSecond: true,
  rounding: 'floor',
  updateInterval: 1000,
}

// Add Column dialog state
const openAddColumn = ref(false)
const newColumnTitle = ref('')

function handleAddColumn() {
  const title = newColumnTitle.value.trim()
  if (!title)
    return
  // prevent duplicate by title (case-insensitive)
  const exists = board.value.columns.some(c => c.title.trim().toLowerCase() === title.toLowerCase())
  if (exists)
    return
  addColumn(title)
  newColumnTitle.value = ''
  openAddColumn.value = false
}
// Expose helpers for parents (e.g., header controls)
function addColumnExternal(title: string) {
  const trimmed = title.trim()
  if (!trimmed)
    return
  // prevent duplicates by title
  const exists = board.value.columns.some(c => c.title.trim().toLowerCase() === trimmed.toLowerCase())
  if (exists)
    return
  addColumn(trimmed)
}

// Completion helpers
function isDoneColumn(colIdOrObj: string | { id: string, title: string }) {
  const id = typeof colIdOrObj === 'string' ? colIdOrObj : colIdOrObj.id
  const title = typeof colIdOrObj === 'string'
    ? (board.value.columns.find(c => c.id === colIdOrObj)?.title || '')
    : colIdOrObj.title
  const t = String(title).trim().toLowerCase()
  return id === 'done' || t === 'done' || t.includes('done') || t === 'completed' || t.includes('complete')
}

const taskToColumnMap = computed<Record<string, string>>(() => {
  const map: Record<string, string> = {}
  for (const col of board.value.columns) {
    for (const t of col.tasks) map[t.id] = col.id
  }
  return map
})

function childrenOf(taskId: string) {
  return board.value.columns.flatMap(c => c.tasks).filter(t => t.parentId === taskId)
}

function taskCompletion(t: Task): number {
  // If task is in a Done-like column, it's 100%
  const colId = taskToColumnMap.value[t.id]
  if (colId) {
    const col = board.value.columns.find(c => c.id === colId)
    if (col && isDoneColumn(col))
      return 100
  }
  // If it has child tasks, compute completion from children in Done columns
  const kids = childrenOf(t.id)
  if (kids.length) {
    const doneKids = kids.filter((kt) => {
      const kcolId = taskToColumnMap.value[kt.id]
      if (!kcolId)
        return false
      const kcol = board.value.columns.find(c => c.id === kcolId)
      return !!kcol && isDoneColumn(kcol)
    }).length
    return Math.round((doneKids / kids.length) * 100)
  }
  // Else if it has subtasks with done flags, compute by them
  if (t.subtasks && t.subtasks.length) {
    const done = t.subtasks.filter(s => s.done).length
    return Math.round((done / t.subtasks.length) * 100)
  }
  // Otherwise 0%
  return 0
}

defineExpose({ addColumnExternal, board })
</script>

<template>
  <div class="flex gap-4 overflow-x-auto overflow-y-hidden pb-4">
    <!-- Columns Draggable wrapper -->
    <Draggable
      v-model="board.columns"
      class="flex gap-4 min-w-max"
      item-key="id"
      :animation="180"
      handle=".col-handle"
      ghost-class="opacity-50"
      @end="onColumnDrop"
    >
      <template #item="{ element: col }: { element: Column }">
        <Card class="w-[272px] shrink-0 py-2 gap-4 self-start">
          <CardHeader class="flex flex-row items-center justify-between gap-2 px-2">
            <CardTitle class="font-semibold text-base flex items-center gap-2">
              <Icon name="lucide:grip-vertical" class="col-handle cursor-grab opacity-60" />
              <span
                :id="`col-title-${col.id}`"
                contenteditable="true" class="hover:bg-accent hover:text-accent-foreground dark:hover:bg-accent/50 px-1 rounded"
                @blur="onUpdateColumn($event, col.id)" @keydown.enter.prevent
              >{{ col.title }}</span>
              <Badge variant="secondary" class="h-5 min-w-5 px-1 font-mono tabular-nums">
                {{ col.tasks.length }}
              </Badge>
            </CardTitle>
            <CardAction class="flex">
              <Button size="icon-sm" variant="ghost" class="size-7 text-muted-foreground" @click="openNewTask(col.id)">
                <Icon name="lucide:plus" />
              </Button>
              <DropdownMenu modal>
                <DropdownMenuTrigger as-child>
                  <Button size="icon-sm" variant="ghost" class="size-7 text-muted-foreground">
                    <Icon name="lucide:ellipsis-vertical" />
                  </Button>
                </DropdownMenuTrigger>
                <DropdownMenuContent class="w-20" align="start">
                  <DropdownMenuItem @click="renameColumn(col.id)">
                    <Icon name="lucide:edit-2" class="size-4" />
                    Rename
                  </DropdownMenuItem>
                  <DropdownMenuSeparator />
                  <DropdownMenuItem variant="destructive" class="text-destructive" @click="removeColumn(col.id)">
                    <Icon name="lucide:trash-2" class="size-4" />
                    Delete
                  </DropdownMenuItem>
                </DropdownMenuContent>
              </DropdownMenu>
            </CardAction>
          </CardHeader>
          <CardContent class="px-2 overflow-y-auto overflow-x-hidden flex-1">
            <!-- Tasks within the column -->
            <Draggable
              v-model="col.tasks"
              :group="{ name: 'kanban-tasks', pull: true, put: true }"
              item-key="id"
              :animation="180"
              class="flex flex-col gap-3 min-h-[24px] p-0.5"
              ghost-class="opacity-50"
              @end="onTaskDrop"
            >
              <template #item="{ element: t }: { element: Task }">
                <div class="rounded-xl border bg-card px-3 py-2 shadow-sm hover:bg-accent/50 cursor-pointer" @click="showEditTask(col.id, t.id)">
                  <div class="flex items-start justify-between gap-2">
                    <div class="text-sm text-muted-foreground">
                      {{ t.id }}
                    </div>
                    <DropdownMenu>
                      <DropdownMenuTrigger as-child>
                        <Button size="icon-sm" variant="ghost" class="size-7 text-muted-foreground" title="More actions" @click.stop>
                          <Icon name="lucide:ellipsis-vertical" />
                        </Button>
                      </DropdownMenuTrigger>
                      <DropdownMenuContent class="w-20" align="start">
                        <DropdownMenuItem @click="showEditTask(col.id, t.id)">
                          <Icon name="lucide:edit-2" class="size-4" />
                          Edit
                        </DropdownMenuItem>
                        <DropdownMenuSeparator />
                        <DropdownMenuItem>
                          <Icon name="lucide:copy" class="size-4" />
                          Copy card
                        </DropdownMenuItem>
                        <DropdownMenuItem>
                          <Icon name="lucide:link" class="size-4" />
                          Copy link
                        </DropdownMenuItem>
                        <DropdownMenuSeparator />
                        <DropdownMenuItem variant="destructive" class="text-destructive" @click="removeTask(col.id, t.id)">
                          <Icon name="lucide:trash-2" class="size-4" />
                          Delete
                        </DropdownMenuItem>
                      </DropdownMenuContent>
                    </DropdownMenu>
                  </div>
                  <p class="font-medium leading-5 mt-1">
                    {{ t.title }}
                  </p>
                  <div v-if="t.labels?.length" class="mt-3 flex items-center gap-1.5">
                    <Badge v-for="lb in t.labels" :key="lb" variant="outline">{{ lb }}</Badge>
                  </div>
                  <div class="mt-3 flex items-center justify-between gap-2">
                    <div class="flex items-center gap-2">
                      <div class="flex items-center text-sm text-muted-foreground gap-1">
                        <Icon name="lucide:list-checks" />
                        <template v-if="t.subtasks?.length">
                          <span>{{ t.subtasks.filter(s => s.done).length }}/{{ t.subtasks.length }}</span>
                        </template>
                        <span class="ml-1 text-[11px]">({{ taskCompletion(t) }}%)</span>
                      </div>
                      <div v-if="t.attachments?.length" class="flex items-center text-sm text-muted-foreground gap-1">
                        <Icon name="lucide:paperclip" />
                        <span>{{ t.attachments.length }}</span>
                      </div>
                      <div v-if="t.comments?.length" class="flex items-center text-sm text-muted-foreground gap-1">
                        <Icon name="lucide:message-square" />
                        <span>{{ t.comments.length }}</span>
                      </div>
                      <div class="flex items-center text-sm text-muted-foreground gap-1">
                        <Icon name="lucide:clock-fading" />
                        <span>{{ useTimeAgo(t.dueDate ?? '', OPTIONS) }}</span>
                      </div>
                      <div v-if="t.parentId" class="flex items-center text-xs text-muted-foreground gap-1">
                        <Icon name="lucide:git-branch" />
                        <span>{{ t.parentId }}</span>
                      </div>
                    </div>
                    <div class="flex items-center gap-2">
                      <Tooltip>
                        <TooltipTrigger as-child>
                          <Icon v-if="t.priority" :name="iconPriority(t.priority)" class="size-4" :class="colorPriority(t.priority)" />
                        </TooltipTrigger>
                        <TooltipContent class="capitalize">
                          {{ t.priority }}
                        </TooltipContent>
                      </Tooltip>
                      <div v-if="t.assignees?.length" class="flex -space-x-2">
                        <Avatar v-for="a in t.assignees.slice(0, 3)" :key="a.id" class="size-6 ring-2 ring-background">
                          <AvatarImage :src="a.avatar" :alt="a.name" />
                          <AvatarFallback class="text-[10px]">
                            {{ a.name.split(' ').map(s => s[0]).slice(0, 2).join('') }}
                          </AvatarFallback>
                        </Avatar>
                      </div>
                    </div>
                  </div>
                </div>
              </template>
            </Draggable>
          </CardContent>
          <CardFooter class="px-2 mt-auto">
            <Button size="sm" variant="ghost" class="text-muted-foreground" @click="openNewTask(col.id)">
              <Icon name="lucide:plus" />
              Add Task
            </Button>
          </CardFooter>
        </Card>
      </template>
    </Draggable>
    <!-- Add Column Card (only on standalone Kanban, hidden on Projects boards) -->
    <Card v-if="!props.projectId" class="w-[272px] shrink-0 py-2 gap-4 self-start border-dashed">
      <div class="h-full flex items-center justify-center p-4">
        <Button variant="outline" class="w-full" @click="openAddColumn = true">
          <Icon name="lucide:plus" class="mr-1" /> Add Column
        </Button>
      </div>
    </Card>
  </div>

  <!-- New Task Dialog -->
  <Dialog v-model:open="showModalTask.open">
    <DialogContent class="sm:max-w-[520px]">
      <DialogHeader>
        <DialogTitle>{{ showModalTask.type === 'create' ? 'New Task' : 'Edit Task' }}</DialogTitle>
        <DialogDescription class="sr-only">
          {{ showModalTask.type === 'create' ? 'Add a new task to the board' : 'Edit the task' }}
        </DialogDescription>
      </DialogHeader>
      <div class="flex flex-col gap-3">
        <div class="grid items-baseline grid-cols-1 md:grid-cols-4 md:[&>label]:col-span-1 *:col-span-3 gap-3">
          <Label>Title</Label>
          <Input v-model="newTask.title" placeholder="Title" />
          <Label>Description</Label>
          <Textarea v-model="newTask.description" placeholder="Description (optional)" rows="4" />
          <Label>Status</Label>
          <Select v-model="(newTask as any).status">
            <SelectTrigger class="w-full">
              <SelectValue placeholder="Select status/column" />
            </SelectTrigger>
            <SelectContent>
              <SelectItem v-for="c in board.columns" :key="c.id" :value="c.id as any">{{ c.title }}</SelectItem>
            </SelectContent>
          </Select>
          <Label>Parent Task</Label>
          <Select v-model="(newTask as any).parentId">
            <SelectTrigger class="w-full">
              <SelectValue placeholder="No parent" />
            </SelectTrigger>
            <SelectContent>
              <SelectItem :value="null as any">None</SelectItem>
              <SelectItem
                v-for="opt in board.columns.flatMap(c => c.tasks).filter(t => t.id !== showModalTask.taskId)"
                :key="opt.id"
                :value="opt.id as any"
              >{{ opt.id }} — {{ opt.title }}</SelectItem>
            </SelectContent>
          </Select>
          <Label>Assignees</Label>
          <DropdownMenu>
            <DropdownMenuTrigger as-child>
              <Button variant="outline" class="justify-start">
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
              <div class="px-2 py-1.5 text-xs text-muted-foreground">Staff</div>
              <DropdownMenuSeparator />
              <DropdownMenuCheckboxItem
                v-for="m in members"
                :key="m.id"
                :checked="assigneeIds.includes(m.id)"
                @update:checked="(v: boolean) => { const set = new Set(assigneeIds.value); v ? set.add(m.id) : set.delete(m.id); assigneeIds.value = Array.from(set) }"
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
          <Label>Priority</Label>
          <Select v-model="newTask.priority">
            <SelectTrigger class="w-full">
              <SelectValue placeholder="Select a priority" />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value="low">
                Low
              </SelectItem>
              <SelectItem value="medium">
                Medium
              </SelectItem>
              <SelectItem value="high">
                High
              </SelectItem>
            </SelectContent>
          </Select>
          <Label>Start Date</Label>
          <div>
            <Popover>
              <PopoverTrigger as-child>
                <Button
                  variant="outline"
                  :class="cn(
                    'w-full justify-start text-left font-normal px-3',
                    !startDate && 'text-muted-foreground',
                  )"
                >
                  <Icon name="lucide:calendar" class="mr-2" />
                  {{ startDate ? df.format(startDate.toDate(getLocalTimeZone())) : "Pick a start date" }}
                </Button>
              </PopoverTrigger>
              <PopoverContent class="w-auto p-0">
                <Calendar v-model="startDate" initial-focus />
              </PopoverContent>
            </Popover>
          </div>
          <Label>Due Date</Label>
          <div class="flex items-center gap-1">
            <Popover>
              <PopoverTrigger as-child>
                <Button
                  variant="outline"
                  :class="cn(
                    'flex-1 justify-start text-left font-normal px-3',
                    !dueDate && 'text-muted-foreground',
                  )"
                >
                  <Icon name="lucide:calendar" class="mr-2" />
                  {{ dueDate ? df.format(dueDate.toDate(getLocalTimeZone())) : "Pick a date" }}
                </Button>
              </PopoverTrigger>
              <PopoverContent class="w-auto p-0">
                <Calendar v-model="dueDate" initial-focus />
              </PopoverContent>
            </Popover>
            <Input
              id="time-picker"
              v-model="dueTime"
              type="time"
              step="60"
              default-value="00:00"
              class="flex-1 bg-background appearance-none [&::-webkit-calendar-picker-indicator]:hidden [&::-webkit-calendar-picker-indicator]:appearance-none"
            />
          </div>
          <Label class="flex items-center justify-between">
            <span>Subtasks</span>
            <Button size="xs" variant="outline" @click="openCreateSubtask"><Icon name="lucide:plus" /> Add subtask</Button>
          </Label>
          <div class="rounded-md border p-2 max-h-40 overflow-y-auto space-y-2">
            <div v-if="subtaskItems.length === 0" class="text-xs text-muted-foreground">
              No subtasks yet. Click "Add subtask" to create one.
            </div>
            <div v-for="(s, i) in subtaskItems" :key="s.id" class="flex items-center gap-2 group">
              <Checkbox :checked="s.done" @update:checked="(v: boolean) => toggleSubtask(i, v)" />
              <button class="text-sm text-left flex-1 truncate hover:underline" @click="openEditSubtask(i)">
                {{ s.title }}
              </button>
              <Button size="icon-sm" variant="ghost" class="size-7 opacity-0 group-hover:opacity-100" title="Edit" @click="openEditSubtask(i)">
                <Icon name="lucide:edit-2" />
              </Button>
              <Button size="icon-sm" variant="ghost" class="size-7 opacity-0 group-hover:opacity-100 text-destructive" title="Delete" @click="removeSubtask(i)">
                <Icon name="lucide:trash-2" />
              </Button>
            </div>
          </div>
          <Label>Attachments</Label>
          <div class="space-y-2">
            <div class="flex items-center gap-2">
              <Button variant="outline" size="sm" @click="fileInput?.click()"><Icon name="i-lucide-paperclip" /> Add Files</Button>
              <input ref="fileInput" type="file" class="hidden" multiple @change="(e: any) => onFilesSelected(e)" />
            </div>
            <div v-if="attachments.length" class="flex flex-wrap gap-2">
              <div v-for="att in attachments" :key="att.id" class="rounded border p-1">
                <img v-if="att.isImage && att.url" :src="att.url" :alt="att.name" class="max-h-16 object-contain" />
                <div v-else class="text-sm flex items-center gap-2"><Icon name="i-lucide-file" /> {{ att.name }}</div>
              </div>
            </div>
          </div>
          <Label>Comments</Label>
          <div class="rounded-md border p-2 max-h-40 overflow-y-auto space-y-2">
            <div v-for="c in comments" :key="c.id" class="text-sm">
              <span class="font-medium">{{ c.author.name }}</span>
              <span class="text-xs text-muted-foreground"> • {{ new Date(c.createdAt).toLocaleString() }}</span>
              <p>{{ c.body }}</p>
            </div>
            <div v-if="comments.length === 0" class="text-xs text-muted-foreground">No comments yet.</div>
          </div>
          <div class="col-span-3 md:col-start-2 flex items-end gap-2">
            <Textarea v-model="commentDraft" placeholder="Write a comment..." rows="2" />
            <Button size="sm" :disabled="!commentDraft.trim()" @click="addComment"><Icon name="i-lucide-send" /> Add</Button>
          </div>
        </div>
      </div>
      <DialogFooter>
        <Button variant="secondary" @click="showModalTask.open = false">
          Cancel
        </Button>
        <Button @click="showModalTask.type === 'create' ? createTask() : editTask()">
          {{ showModalTask.type === 'create' ? 'Create' : 'Update' }}
        </Button>
      </DialogFooter>
    </DialogContent>
  </Dialog>

  <!-- Subtask Create/Edit Dialog -->
  <Dialog v-model:open="openSubtask">
    <DialogContent class="sm:max-w-[420px]">
      <DialogHeader>
        <DialogTitle>{{ subtaskMode === 'create' ? 'Add Subtask' : 'Edit Subtask' }}</DialogTitle>
      </DialogHeader>
      <div class="space-y-3 py-2">
        <div class="space-y-2">
          <Label for="stitle">Title</Label>
          <Input id="stitle" v-model="subtaskForm.title" placeholder="Describe the subtask" />
        </div>
      </div>
      <DialogFooter>
        <Button variant="secondary" @click="openSubtask = false">Cancel</Button>
        <Button :disabled="!subtaskForm.title.trim()" @click="saveSubtask">Save</Button>
      </DialogFooter>
    </DialogContent>
  </Dialog>

  <!-- Add Column Dialog -->
  <Dialog v-model:open="openAddColumn">
    <DialogContent class="sm:max-w-[420px]">
      <DialogHeader>
        <DialogTitle>New Column</DialogTitle>
        <DialogDescription>Add a new column to this board.</DialogDescription>
      </DialogHeader>
      <div class="space-y-3 py-2">
        <div class="space-y-2">
          <Label for="colTitle">Title</Label>
          <Input id="colTitle" v-model="newColumnTitle" placeholder="e.g. In Review" />
        </div>
      </div>
      <DialogFooter>
        <Button variant="secondary" @click="openAddColumn = false">Cancel</Button>
        <Button :disabled="!newColumnTitle.trim()" @click="handleAddColumn">Add</Button>
      </DialogFooter>
    </DialogContent>
  </Dialog>
</template>
