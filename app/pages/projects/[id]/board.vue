<script setup lang="ts">
import { useRoute, useRouter } from 'vue-router'
import KanbanBoard from '~/components/kanban/KanbanBoard.vue'
import { useProjects } from '~/composables/useProjects'

const route = useRoute()
const router = useRouter()
const { getProject, updateProject } = useProjects()

const pid = computed(() => String(route.params.id))
const project = computed(() => getProject(pid.value))

// Boards
const selectedBoard = ref<string>('Board')
watchEffect(() => {
  if (project.value) {
    const boards = project.value.boards && project.value.boards.length ? project.value.boards : ['Board']
    if (!boards.includes(selectedBoard.value))
      selectedBoard.value = boards[0]
  }
})

watchEffect(() => {
  if (!project.value) {
    // If project not found, go back to projects list
    router.replace('/projects')
  }
})

// Settings dialog state
const openSettings = ref(false)
const form = reactive({ name: '', key: '', customerName: '' })

watch(() => openSettings.value, (open) => {
  if (open && project.value) {
    form.name = project.value.name
    form.key = project.value.key
    form.customerName = project.value.customerName || ''
  }
})

function saveSettings() {
  if (!project.value)
    return
  const name = form.name.trim()
  if (!name)
    return
  updateProject(project.value.id, { name, key: form.key, customerName: form.customerName })
  openSettings.value = false
}

const boardRef = ref()
const showNewColumn = ref(false)
const newColumnTitle = ref('')
const newBoardName = ref('')

function createColumn() {
  const title = newColumnTitle.value.trim()
  if (!title)
    return
  // Proxy to KanbanBoard via event
  const el = boardRef.value as any
  if (el && typeof el.exposed?.addColumnExternal === 'function')
    el.exposed.addColumnExternal(title)
  newColumnTitle.value = ''
  showNewColumn.value = false
}

function createBoard() {
  const name = newBoardName.value.trim()
  if (!project.value || !name)
    return
  const boards = Array.from(new Set([...(project.value.boards || ['Board']), name]))
  updateProject(project.value.id, { boards })
  selectedBoard.value = name
  newBoardName.value = ''
}

// --- Completion percentage logic (overall project) ---
const projectCompletion = reactive({ done: 0, total: 0, percent: 0 })

function isDoneTitle(title?: string) {
  const t = String(title || '').trim().toLowerCase()
  return t === 'done' || t.includes('done') || t === 'completed' || t.includes('complete')
}

function recalcOverall() {
  const p = project.value
  if (!p)
    return
  let done = 0
  let total = 0
  const boards = p.boards && p.boards.length ? p.boards : ['Board']

  // Helper to count for a board state
  const countFromState = (state: any) => {
    if (!state || !Array.isArray(state.columns))
      return { d: 0, t: 0 }
    let d = 0
    let t = 0
    for (const col of state.columns) {
      const list = Array.isArray(col.tasks) ? col.tasks : []
      t += list.length
      if (isDoneTitle(col.title) || col.id === 'done')
        d += list.length
    }
    return { d, t }
  }

  const STORAGE_KEY = 'kanban.board.v1'
  for (const b of boards) {
    if (b === selectedBoard.value && (boardRef as any)?.value?.exposed?.board) {
      const state = (boardRef as any).value.exposed.board.value
      const { d, t } = countFromState(state)
      done += d
      total += t
    }
    else if (import.meta.client) {
      const raw = localStorage.getItem(`${STORAGE_KEY}:${pid.value}:${b}`) || localStorage.getItem(`${STORAGE_KEY}:${pid.value}`)
      if (raw) {
        try {
          const parsed = JSON.parse(raw)
          const { d, t } = countFromState(parsed)
          done += d
          total += t
        }
        catch {}
      }
    }
  }
  projectCompletion.done = done
  projectCompletion.total = total
  projectCompletion.percent = total > 0 ? Math.round((done / total) * 100) : 0
}

// Recompute on dependency changes
watch([selectedBoard, () => project.value?.boards], () => recalcOverall(), { immediate: true })
onMounted(() => {
  recalcOverall()
  // Listen to storage changes from other tabs
  if (import.meta.client) {
    window.addEventListener('storage', recalcOverall)
  }
})
onBeforeUnmount(() => {
  if (import.meta.client) {
    window.removeEventListener('storage', recalcOverall)
  }
})
// Watch current board state for reactive updates
watchEffect(() => {
  // access columns/tasks to create a reactive dependency
  const cols = (boardRef as any)?.value?.exposed?.board?.value?.columns
  if (cols) {
    // iterate to depend

    cols.map((c: any) => c.tasks.length)
    recalcOverall()
  }
})
</script>

<template>
  <div class="-m-4 lg:-m-6 h-full flex flex-col">
    <!-- Header -->
    <div class="flex h-[56px] items-center gap-2 border-b px-4 justify-between">
      <div class="flex items-center gap-3 min-w-0">
        <Button size="icon" variant="ghost" title="Back" @click="router.push('/projects')">
          <Icon name="i-lucide-arrow-left" />
        </Button>
        <h2 class="text-xl font-semibold truncate">
          {{ project?.name || 'Project' }}
        </h2>
        <Badge v-if="project?.key" variant="secondary">
          {{ project?.key }}
        </Badge>
        <span v-if="project?.customerName" class="text-sm text-muted-foreground truncate">• {{ project?.customerName }}</span>
      </div>
      <div class="flex items-center gap-2">
        <!-- Overall Project Completion -->
        <div v-if="projectCompletion.total > 0" class="hidden sm:flex items-center gap-2 mr-2">
          <div class="text-xs text-muted-foreground">
            Project
          </div>
          <Badge variant="secondary" class="font-mono">
            {{ projectCompletion.percent }}%
          </Badge>
        </div>
        <div class="flex items-center gap-2">
          <Label class="sr-only">Board</Label>
          <Select v-model="selectedBoard">
            <SelectTrigger class="w-[160px]">
              <SelectValue placeholder="Select board" />
            </SelectTrigger>
            <SelectContent>
              <SelectItem v-for="b in (project?.boards || ['Board'])" :key="b" :value="b">
                {{ b }}
              </SelectItem>
            </SelectContent>
          </Select>
          <Dialog>
            <DialogTrigger as-child>
              <Button size="icon-sm" variant="outline" title="Add board">
                <Icon name="i-lucide-plus" />
              </Button>
            </DialogTrigger>
            <DialogContent class="sm:max-w-[420px]">
              <DialogHeader>
                <DialogTitle>New Board</DialogTitle>
              </DialogHeader>
              <div class="space-y-3 py-2">
                <div class="space-y-2">
                  <Label for="bname">Board name</Label>
                  <Input id="bname" v-model="newBoardName" placeholder="e.g. Development" />
                </div>
                <div class="flex items-center gap-2">
                  <DialogClose as-child>
                    <Button variant="secondary">
                      Cancel
                    </Button>
                  </DialogClose>
                  <DialogClose as-child>
                    <Button :disabled="!newBoardName.trim()" @click="createBoard">
                      Create
                    </Button>
                  </DialogClose>
                </div>
              </div>
            </DialogContent>
          </Dialog>
        </div>
        <Button size="sm" variant="outline" @click="router.push(`/projects/${pid}/gantt`)">
          <Icon name="lucide:calendar-range" /> Gantt
        </Button>
        <Button size="sm" @click="showNewColumn = true">
          <Icon name="lucide:plus" />
          Add Column
        </Button>
        <Button size="sm" variant="ghost" @click="openSettings = true">
          <Icon name="i-lucide-settings-2" class="mr-1" /> Settings
        </Button>
        <Button size="sm" variant="outline" @click="router.push('/projects')">
          All Projects
        </Button>
      </div>
    </div>

    <!-- Board content -->
    <section class="flex-1 min-h-0 overflow-hidden p-4">
      <KanbanBoard ref="boardRef" :project-id="pid" :board-id="selectedBoard" />
    </section>

    <!-- Settings Dialog -->
    <Dialog v-model:open="openSettings">
      <DialogContent class="sm:max-w-[480px]">
        <DialogHeader>
          <DialogTitle>Project Settings</DialogTitle>
          <DialogDescription>Update the project name, key and customer/company.</DialogDescription>
        </DialogHeader>
        <div class="space-y-3 py-2">
          <div class="space-y-2">
            <Label for="pname">Name</Label>
            <Input id="pname" v-model="form.name" placeholder="Project name" />
          </div>
          <div class="space-y-2">
            <Label for="pkey">Key</Label>
            <Input id="pkey" v-model="form.key" placeholder="KEY" />
            <p class="text-xs text-muted-foreground">
              Short uppercase code (max 10 chars). Spaces will be replaced with dashes.
            </p>
          </div>
          <div class="space-y-2">
            <Label for="pcustomer">Customer / Company</Label>
            <Input id="pcustomer" v-model="form.customerName" placeholder="Acme Corp" />
          </div>
        </div>
        <DialogFooter>
          <Button variant="secondary" @click="openSettings = false">
            Cancel
          </Button>
          <Button :disabled="!form.name.trim()" @click="saveSettings">
            Save
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>

    <!-- New Column Dialog (matches standalone Kanban page) -->
    <Dialog v-model:open="showNewColumn">
      <DialogContent class="sm:max-w-[420px]">
        <DialogHeader>
          <DialogTitle>Add New Column</DialogTitle>
          <DialogDescription class="sr-only">
            Add a new column to the board
          </DialogDescription>
        </DialogHeader>
        <form name="newColumnForm" class="flex flex-col gap-3" @submit.prevent="createColumn">
          <Input v-model="newColumnTitle" placeholder="Column title" />
        </form>
        <DialogFooter>
          <Button variant="secondary" @click="showNewColumn = false">
            Cancel
          </Button>
          <Button type="submit" form="newColumnForm" @click="createColumn">
            Create
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  </div>
</template>
