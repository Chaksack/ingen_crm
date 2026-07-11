<script setup lang="ts">
import { h } from 'vue'
import type { ColumnDef } from '@tanstack/vue-table'

interface Queue { id: string; name: string; routing_strategy: string }
interface OrgMember { id: string; display_name: string; email: string }
interface QueueMember { id: string; display_name: string; email: string }

const api = useApi()
const auth = useAuthStore()

const queues = ref<Queue[]>([])
const newName = ref('')
const creating = ref(false)

const routingOptions = [
  { value: 'manual', label: 'Manual' },
  { value: 'round_robin', label: 'Round robin' },
  { value: 'capacity', label: 'Least busy (capacity)' },
]

const membersDialogOpen = ref(false)
const activeQueue = ref<Queue | null>(null)
const orgMembers = ref<OrgMember[]>([])
const queueMembers = ref<QueueMember[]>([])
const addMemberId = ref('')

async function load() {
  queues.value = await api<Queue[]>('/service/queues') ?? []
}

async function createQueue() {
  if (!newName.value.trim()) return
  creating.value = true
  try {
    await api('/service/queues', { method: 'POST', body: { name: newName.value } })
    newName.value = ''
    await load()
  } finally {
    creating.value = false
  }
}

async function setRouting(queue: Queue, strategy: string) {
  await api(`/service/queues/${queue.id}/routing`, { method: 'PATCH', body: { routing_strategy: strategy } })
  await load()
}

async function openMembers(queue: Queue) {
  activeQueue.value = queue
  membersDialogOpen.value = true
  addMemberId.value = ''
  const [org, members] = await Promise.all([
    api<OrgMember[]>('/team/members'),
    api<QueueMember[]>(`/service/queues/${queue.id}/members`),
  ])
  orgMembers.value = org ?? []
  queueMembers.value = members ?? []
}

async function addMember() {
  if (!addMemberId.value || !activeQueue.value) return
  await api(`/service/queues/${activeQueue.value.id}/members`, { method: 'POST', body: { user_id: addMemberId.value } })
  queueMembers.value = await api<QueueMember[]>(`/service/queues/${activeQueue.value.id}/members`) ?? []
  addMemberId.value = ''
}

async function removeMember(userId: string) {
  if (!activeQueue.value) return
  await api(`/service/queues/${activeQueue.value.id}/members/${userId}`, { method: 'DELETE' })
  queueMembers.value = await api<QueueMember[]>(`/service/queues/${activeQueue.value.id}/members`) ?? []
}

const availableToAdd = computed(() =>
  orgMembers.value.filter((m) => !queueMembers.value.some((qm) => qm.id === m.id)),
)

const columns = computed<ColumnDef<Queue, any>[]>(() => {
  const base: ColumnDef<Queue, any>[] = [
    { accessorKey: 'name', header: 'Name', enableSorting: true },
  ]
  if (!auth.isAdmin) return base
  return [
    ...base,
    {
      id: 'routing',
      header: 'Routing',
      cell: ({ row }) => h(
        resolveComponent('Select'),
        {
          modelValue: row.original.routing_strategy,
          'onUpdate:modelValue': (v: string) => setRouting(row.original, v),
        },
        {
          default: () => [
            h(resolveComponent('SelectTrigger'), { class: 'w-48' }, { default: () => h(resolveComponent('SelectValue')) }),
            h(resolveComponent('SelectContent'), {}, {
              default: () => routingOptions.map((opt) => h(resolveComponent('SelectItem'), { value: opt.value, key: opt.value }, () => opt.label)),
            }),
          ],
        },
      ),
    },
    {
      id: 'actions',
      header: '',
      cell: ({ row }) => h(resolveComponent('Button'), { variant: 'ghost', size: 'sm', onClick: () => openMembers(row.original) }, () => 'Manage members'),
    },
  ]
})

onMounted(load)
</script>

<template>
  <div class="flex flex-col gap-6">
    <div>
      <h1 class="text-2xl font-semibold tracking-tight">Queues</h1>
      <p class="text-muted-foreground">Case routing destinations, typically one per client engagement.</p>
    </div>

    <Card>
      <CardHeader>
        <CardTitle class="text-base">New queue</CardTitle>
      </CardHeader>
      <CardContent>
        <form class="flex gap-2" @submit.prevent="createQueue">
          <Input v-model="newName" placeholder="Queue name" class="max-w-sm" />
          <Button type="submit" :disabled="creating">Add</Button>
        </form>
      </CardContent>
    </Card>

    <BaseDataTable :columns="columns" :data="queues" search-placeholder="Filter queues…" />

    <Dialog v-model:open="membersDialogOpen">
      <DialogContent>
        <DialogHeader>
          <DialogTitle>{{ activeQueue?.name }} members</DialogTitle>
          <DialogDescription>Agents eligible for auto-assignment in this queue.</DialogDescription>
        </DialogHeader>
        <div class="space-y-4">
          <div class="flex gap-2">
            <Select v-model="addMemberId">
              <SelectTrigger><SelectValue placeholder="Add a teammate…" /></SelectTrigger>
              <SelectContent>
                <SelectItem v-for="m in availableToAdd" :key="m.id" :value="m.id">{{ m.display_name }}</SelectItem>
              </SelectContent>
            </Select>
            <Button :disabled="!addMemberId" @click="addMember">Add</Button>
          </div>
          <ul class="space-y-2">
            <li v-for="m in queueMembers" :key="m.id" class="flex items-center justify-between rounded-md border border-border px-3 py-2 text-sm">
              <span>{{ m.display_name }} <span class="text-muted-foreground">({{ m.email }})</span></span>
              <Button variant="ghost" size="sm" @click="removeMember(m.id)">Remove</Button>
            </li>
            <li v-if="queueMembers.length === 0" class="text-sm text-muted-foreground">No members yet.</li>
          </ul>
        </div>
      </DialogContent>
    </Dialog>
  </div>
</template>
