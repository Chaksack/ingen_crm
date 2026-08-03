<script setup lang="ts">
import type { Task } from '~/types/kanban'
import { useKanban } from '~/composables/useKanban'
import { useProjects } from '~/composables/useProjects'

const props = withDefaults(defineProps<{ projectId: string, boardId?: string }>(), {
  boardId: 'default',
})

const { board } = useKanban(props.projectId, props.boardId)
const { getProject } = useProjects()
const project = computed(() => getProject(props.projectId))

const tasks = computed<Task[]>(() => board.value.columns.flatMap(c => c.tasks))

function toDate(val?: any): Date | undefined {
  if (!val)
    return undefined
  if (val instanceof Date)
    return val
  const d = new Date(val)
  return Number.isNaN(d.getTime()) ? undefined : d
}

const dated = computed(() => tasks.value
  .map(t => ({
    ...t,
    _start: toDate((t as any).startDate),
    _due: toDate(t.dueDate),
  }))
  .filter(t => t._start && t._due) as Array<Task & { _start: Date, _due: Date }>)

const minStart = computed(() => {
  const m = dated.value.reduce<Date | undefined>((acc, t) => !acc || t._start < acc ? t._start : acc, undefined as any)
  return m || new Date()
})
const maxDue = computed(() => {
  const m = dated.value.reduce<Date | undefined>((acc, t) => !acc || t._due > acc ? t._due : acc, undefined as any)
  return m || new Date()
})

const daysSpan = computed(() => {
  const ms = Math.max(1, Math.ceil((maxDue.value.getTime() - minStart.value.getTime()) / (1000 * 60 * 60 * 24)))
  return ms + 1
})

const days = computed(() => Array.from({ length: daysSpan.value }, (_, i) => {
  const d = new Date(minStart.value)
  d.setDate(d.getDate() + i)
  return d
}))

function dayIndex(date: Date) {
  const start = new Date(minStart.value)
  start.setHours(0, 0, 0, 0)
  const cur = new Date(date)
  cur.setHours(0, 0, 0, 0)
  return Math.max(0, Math.floor((cur.getTime() - start.getTime()) / (1000 * 60 * 60 * 24)))
}

function pct(n: number) {
  return `${n}%`
}

const empty = computed(() => dated.value.length === 0)
</script>

<template>
  <div class="flex flex-col gap-3">
    <div class="flex items-center justify-between">
      <div>
        <h3 class="font-semibold text-lg">Gantt — {{ project?.name }}</h3>
        <p class="text-sm text-muted-foreground" v-if="project?.customerName">Customer: {{ project?.customerName }}</p>
      </div>
      <Badge v-if="project?.key" variant="secondary">{{ project?.key }}</Badge>
    </div>

    <div v-if="empty" class="rounded-md border p-6 text-sm text-muted-foreground">
      No tasks with start and due dates yet. Add tasks with both dates to see the timeline.
    </div>

    <div v-else class="rounded-md border overflow-hidden">
      <!-- Header timeline scale -->
      <div class="grid bg-muted/50 text-xs text-muted-foreground" :style="{ gridTemplateColumns: `200px repeat(${days.length}, 1fr)` }">
        <div class="p-2 font-medium bg-background">Task</div>
        <div v-for="d in days" :key="d.toISOString()" class="p-2 text-center border-l">
          {{ d.toLocaleDateString(undefined, { month: 'short', day: 'numeric' }) }}
        </div>
      </div>
      <!-- Rows -->
      <div class="divide-y">
        <div v-for="t in dated" :key="t.id" class="grid items-stretch" :style="{ gridTemplateColumns: `200px repeat(${days.length}, 1fr)` }">
          <div class="p-2 flex items-center gap-2 min-w-0">
            <div class="shrink-0 text-xs text-muted-foreground">{{ t.id }}</div>
            <div class="truncate">{{ t.title }}</div>
          </div>
          <div class="relative col-span-full">
            <div class="relative h-10">
              <div class="absolute inset-0 grid" :style="{ gridTemplateColumns: `repeat(${days.length}, 1fr)` }">
                <div v-for="i in days.length" :key="i" class="border-l first:border-l-0" />
              </div>
              <div
                class="absolute top-1/2 -translate-y-1/2 h-6 rounded bg-primary/20 border border-primary/40 hover:bg-primary/30 transition-colors"
                :style="{
                  left: pct((dayIndex(t._start) / Math.max(1, days.length)) * 100),
                  width: pct(((dayIndex(t._due) - dayIndex(t._start) + 1) / Math.max(1, days.length)) * 100),
                }"
              >
                <div class="px-2 text-xs text-primary-foreground/90 truncate">
                  {{ t.title }}
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
</style>
