<script setup lang="ts">
import { useRoute, useRouter } from 'vue-router'
import { useProjects } from '~/composables/useProjects'

const route = useRoute()
const router = useRouter()
const { getProject, members } = useProjects()

const pid = computed(() => String(route.params.id))
const project = computed(() => getProject(pid.value))

watchEffect(() => {
  if (!project.value) {
    router.replace('/projects')
  }
})

interface TaskItem {
  id: string
  title: string
  description?: string
  board: string
  columnId: string
  columnTitle: string
  createdAt?: string
}

interface StatusCount {
  key: string
  title: string
  count: number
}

const STORAGE_KEY = 'kanban.board.v1'

function isDoneTitle(title?: string) {
  const t = String(title || '').trim().toLowerCase()
  return t === 'done' || t.includes('done') || t === 'completed' || t.includes('complete')
}

const summary = reactive({
  total: 0,
  done: 0,
  percent: 0,
  byStatus: [] as StatusCount[],
  tasks: [] as TaskItem[],
})

function loadAllBoards() {
  const p = project.value
  if (!p) {
    return
  }
  const boards = p.boards && p.boards.length ? p.boards : ['Board']
  let total = 0
  let done = 0
  const statusMap = new Map<string, { title: string, count: number }>()
  const tasks: TaskItem[] = []

  const countFromState = (state: any, boardName: string) => {
    if (!state || !Array.isArray(state.columns)) {
      return
    }
    for (const col of state.columns) {
      const list = Array.isArray(col.tasks) ? col.tasks : []
      // status counts
      const key = String(col.id || col.title || '').toLowerCase()
      const entry = statusMap.get(key) || { title: String(col.title || col.id || 'Unknown'), count: 0 }
      entry.count += list.length
      statusMap.set(key, entry)
      // totals
      total += list.length
      if (isDoneTitle(col.title) || col.id === 'done') {
        done += list.length
      }
      // task list
      for (const t of list) {
        tasks.push({
          id: String(t.id || ''),
          title: String(t.title || t.name || '(No title)'),
          description: t.description,
          board: boardName,
          columnId: String(col.id || ''),
          columnTitle: String(col.title || ''),
          createdAt: t.createdAt ? new Date(t.createdAt).toLocaleString() : undefined,
        })
      }
    }
  }

  for (const b of boards) {
    let parsed: any = null
    if (import.meta.client) {
      const raw = localStorage.getItem(`${STORAGE_KEY}:${pid.value}:${b}`) || localStorage.getItem(`${STORAGE_KEY}:${pid.value}`)
      if (raw) {
        try {
          parsed = JSON.parse(raw)
        }
        catch {
          // ignore parse error
        }
      }
    }
    if (parsed) {
      countFromState(parsed, b)
    }
  }

  summary.total = total
  summary.done = done
  summary.percent = total > 0 ? Math.round((done / total) * 100) : 0
  summary.byStatus = Array.from(statusMap.entries()).map(([key, v]) => ({ key, title: v.title, count: v.count }))
  // Sort tasks by createdAt desc if present
  summary.tasks = tasks.sort((a, b) => {
    const da = a.createdAt ? Date.parse(a.createdAt) : 0
    const db = b.createdAt ? Date.parse(b.createdAt) : 0
    return db - da
  })
}

onMounted(() => {
  loadAllBoards()
  if (import.meta.client) {
    window.addEventListener('storage', loadAllBoards)
  }
})
onBeforeUnmount(() => {
  if (import.meta.client) {
    window.removeEventListener('storage', loadAllBoards)
  }
})
watch(() => project.value?.boards, () => loadAllBoards(), { immediate: true })

const lead = computed(() => members.value.find(m => m.id === project.value?.leadId))
const projectMembers = computed(() => members.value.filter(m => (project.value?.memberIds || []).includes(m.id)))
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
        <div v-if="summary.total > 0" class="hidden sm:flex items-center gap-2 mr-2">
          <div class="text-xs text-muted-foreground">
            Project
          </div>
          <Badge variant="secondary" class="font-mono">
            {{ summary.percent }}%
          </Badge>
        </div>
        <div class="flex items-center gap-1">
          <Button size="sm" variant="ghost" @click="router.push(`/projects/${pid}/board`)">
            <Icon name="i-lucide-layout-dashboard" /> Board
          </Button>
          <Button size="sm" variant="ghost" @click="router.push(`/projects/${pid}/gantt`)">
            <Icon name="i-lucide-timeline" /> Gantt
          </Button>
        </div>
      </div>
    </div>

    <!-- Content -->
    <div class="p-4 space-y-4 overflow-auto">
      <!-- Project details and completion -->
      <div class="grid grid-cols-1 lg:grid-cols-3 gap-4">
        <Card class="col-span-1 lg:col-span-2">
          <CardHeader>
            <CardTitle>Summary</CardTitle>
            <CardDescription>Overview of project details and progress</CardDescription>
          </CardHeader>
          <CardContent class="space-y-4">
            <div class="grid grid-cols-2 md:grid-cols-4 gap-4">
              <div>
                <div class="text-xs text-muted-foreground">
                  Tasks
                </div>
                <div class="text-xl font-semibold">
                  {{ summary.total }}
                </div>
              </div>
              <div>
                <div class="text-xs text-muted-foreground">
                  Done
                </div>
                <div class="text-xl font-semibold">
                  {{ summary.done }}
                </div>
              </div>
              <div>
                <div class="text-xs text-muted-foreground">
                  Completion
                </div>
                <div class="text-xl font-semibold">
                  {{ summary.percent }}%
                </div>
              </div>
              <div>
                <div class="text-xs text-muted-foreground">
                  Boards
                </div>
                <div class="text-xl font-semibold">
                  {{ (project?.boards || ['Board']).length }}
                </div>
              </div>
            </div>

            <div v-if="summary.byStatus.length" class="flex flex-wrap gap-2">
              <Badge v-for="s in summary.byStatus" :key="s.key" variant="outline">
                {{ s.title }}: {{ s.count }}
              </Badge>
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle>Details</CardTitle>
          </CardHeader>
          <CardContent class="space-y-3">
            <div class="text-sm">
              <span class="text-muted-foreground">Created:</span> {{ project ? new Date(project.createdAt).toLocaleString() : '' }}
            </div>
            <div v-if="lead" class="text-sm">
              <span class="text-muted-foreground">Lead:</span> {{ lead?.name }}
            </div>
            <div v-if="projectMembers.length" class="text-sm">
              <span class="text-muted-foreground">Members:</span>
              <span> {{ projectMembers.map(m => m.name).join(', ') }}</span>
            </div>
          </CardContent>
          <CardFooter class="flex gap-2">
            <Button size="sm" variant="secondary" @click="router.push(`/projects/${pid}/board`)">
              <Icon name="i-lucide-pencil" /> Manage Tasks
            </Button>
          </CardFooter>
        </Card>
      </div>

      <!-- Tasks list -->
      <Card>
        <CardHeader>
          <CardTitle>Tasks</CardTitle>
          <CardDescription>All tasks across boards with their current status</CardDescription>
        </CardHeader>
        <CardContent>
          <div v-if="summary.tasks.length === 0" class="text-sm text-muted-foreground py-6">
            No tasks yet. Use the Board to add tasks.
          </div>
          <div v-else class="divide-y">
            <div v-for="t in summary.tasks" :key="`${t.board}:${t.id}`" class="py-3 flex items-start justify-between gap-3">
              <div class="min-w-0">
                <div class="font-medium truncate">
                  {{ t.title }}
                </div>
                <div v-if="t.description" class="text-xs text-muted-foreground truncate">
                  {{ t.description }}
                </div>
                <div class="text-xs text-muted-foreground">
                  {{ t.id }} • {{ t.board }} • {{ t.createdAt || '' }}
                </div>
              </div>
              <Badge variant="secondary">
                {{ t.columnTitle }}
              </Badge>
            </div>
          </div>
        </CardContent>
      </Card>
    </div>
  </div>
</template>
