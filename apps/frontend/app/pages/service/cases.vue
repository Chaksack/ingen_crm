<script setup lang="ts">
import { h } from 'vue'
import type { ColumnDef } from '@tanstack/vue-table'

interface Queue { id: string; name: string }
interface SLAInfo { status: string; due_at?: string; achieved_at?: string }
interface Case {
  id: string
  subject: string
  priority: string | null
  status: string
  queue_id: string | null
  contact_id: string | null
  first_response_sla: SLAInfo
  resolution_sla: SLAInfo
}

const api = useApi()
const queues = ref<Queue[]>([])
const cases = ref<Case[]>([])
const error = ref('')

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

async function setStatus(item: Case, status: string) {
  error.value = ''
  try {
    await api(`/service/cases/${item.id}`, {
      method: 'PUT',
      body: {
        subject: item.subject, priority: item.priority, status,
        queue_id: item.queue_id, contact_id: item.contact_id,
      },
    })
    await load()
  } catch {
    error.value = `Could not update "${item.subject}" — you may not have permission to edit this case.`
  }
}

async function respond(item: Case) {
  error.value = ''
  try {
    await api(`/service/cases/${item.id}/respond`, { method: 'POST' })
    await load()
  } catch {
    error.value = `Could not respond to "${item.subject}" — you may not have permission to edit this case.`
  }
}

function queueName(id: string | null) {
  return queues.value.find((q) => q.id === id)?.name ?? '—'
}

const statusOptions = ['new', 'open', 'waiting', 'resolved', 'closed']

const slaVariant: Record<string, 'default' | 'outline' | 'secondary' | 'destructive'> = {
  ok: 'outline',
  warning: 'default',
  breached: 'destructive',
  met: 'secondary',
  met_late: 'secondary',
  'n/a': 'outline',
}
const slaLabel: Record<string, string> = {
  ok: 'On track',
  warning: 'Due soon',
  breached: 'Breached',
  met: 'Met',
  met_late: 'Met late',
  'n/a': '—',
}

function slaBadge(sla: SLAInfo) {
  return h(resolveComponent('Badge'), { variant: slaVariant[sla.status] ?? 'outline' }, () => slaLabel[sla.status] ?? sla.status)
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
    id: 'first_response',
    header: 'First response',
    cell: ({ row }) => slaBadge(row.original.first_response_sla),
  },
  {
    id: 'resolution',
    header: 'Resolution',
    cell: ({ row }) => slaBadge(row.original.resolution_sla),
  },
  {
    accessorKey: 'status',
    header: 'Status',
    enableSorting: true,
    cell: ({ row }) => h(
      resolveComponent('Select'),
      {
        modelValue: row.original.status,
        'onUpdate:modelValue': (v: string) => setStatus(row.original, v),
      },
      {
        default: () => [
          h(resolveComponent('SelectTrigger'), { class: 'w-32' }, { default: () => h(resolveComponent('SelectValue')) }),
          h(resolveComponent('SelectContent'), {}, {
            default: () => statusOptions.map((s) => h(resolveComponent('SelectItem'), { value: s, key: s }, () => s)),
          }),
        ],
      },
    ),
  },
  {
    id: 'actions',
    header: '',
    cell: ({ row }) => {
      const item = row.original
      if (item.first_response_sla.status === 'met' || item.first_response_sla.status === 'met_late') return null
      return h(resolveComponent('Button'), { variant: 'ghost', size: 'sm', onClick: () => respond(item) }, () => 'Respond')
    },
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

    <p v-if="error" class="text-sm text-destructive">{{ error }}</p>

    <BaseDataTable :columns="columns" :data="cases" search-placeholder="Filter cases…" />
  </div>
</template>
