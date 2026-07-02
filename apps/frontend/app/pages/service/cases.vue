<script setup lang="ts">
import { h } from 'vue'
import type { ColumnDef } from '@tanstack/vue-table'

interface Queue { id: string; name: string }
interface Case { id: string; subject: string; priority: string | null; status: string; queue_id: string | null }

const api = useApi()
const queues = ref<Queue[]>([])
const cases = ref<Case[]>([])

const newSubject = ref('')
const newPriority = ref('normal')
const newQueueId = ref('none')
const creating = ref(false)

async function load() {
  queues.value = await api<Queue[]>('/service/queues') ?? []
  cases.value = await api<Case[]>('/service/cases') ?? []
}

async function createCase() {
  if (!newSubject.value.trim()) return
  creating.value = true
  try {
    await api('/service/cases', {
      method: 'POST',
      body: {
        subject: newSubject.value,
        priority: newPriority.value,
        queue_id: newQueueId.value === 'none' ? null : newQueueId.value,
      },
    })
    newSubject.value = ''
    await load()
  } finally {
    creating.value = false
  }
}

function queueName(id: string | null) {
  return queues.value.find((q) => q.id === id)?.name ?? '—'
}

const statusVariant: Record<string, 'default' | 'outline' | 'secondary' | 'destructive'> = {
  new: 'default',
  open: 'secondary',
  resolved: 'outline',
}

const columns: ColumnDef<Case, any>[] = [
  { accessorKey: 'subject', header: 'Subject', enableSorting: true },
  { accessorKey: 'priority', header: 'Priority', enableSorting: true },
  {
    id: 'queue',
    header: 'Queue',
    cell: ({ row }) => queueName(row.original.queue_id),
  },
  {
    accessorKey: 'status',
    header: 'Status',
    enableSorting: true,
    cell: ({ row }) => h(resolveComponent('Badge'), { variant: statusVariant[row.original.status] ?? 'default' }, () => row.original.status),
  },
]

onMounted(load)
</script>

<template>
  <div class="flex flex-col gap-6">
    <div>
      <h1 class="text-2xl font-semibold tracking-tight">Cases</h1>
      <p class="text-muted-foreground">Customer support tickets routed through your queues.</p>
    </div>

    <Card>
      <CardHeader>
        <CardTitle class="text-base">New case</CardTitle>
      </CardHeader>
      <CardContent>
        <form class="flex flex-wrap items-center gap-2" @submit.prevent="createCase">
          <Input v-model="newSubject" placeholder="Subject" class="max-w-sm" />
          <Select v-model="newPriority">
            <SelectTrigger class="w-32"><SelectValue /></SelectTrigger>
            <SelectContent>
              <SelectItem value="low">Low</SelectItem>
              <SelectItem value="normal">Normal</SelectItem>
              <SelectItem value="high">High</SelectItem>
            </SelectContent>
          </Select>
          <Select v-model="newQueueId">
            <SelectTrigger class="w-40"><SelectValue placeholder="No queue" /></SelectTrigger>
            <SelectContent>
              <SelectItem value="none">No queue</SelectItem>
              <SelectItem v-for="q in queues" :key="q.id" :value="q.id">{{ q.name }}</SelectItem>
            </SelectContent>
          </Select>
          <Button type="submit" :disabled="creating">Add case</Button>
        </form>
      </CardContent>
    </Card>

    <BaseDataTable :columns="columns" :data="cases" search-placeholder="Filter cases…" />
  </div>
</template>
