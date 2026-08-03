<script setup lang="ts">
import { Mail, MessageCircle, Search } from 'lucide-vue-next'

interface Ticket {
  id: string
  ticketNumber: string
  name: string
  email: string
  company?: string
  subject: string
  priority: 'low' | 'medium' | 'high' | 'urgent'
  status: 'open' | 'in_progress' | 'resolved' | 'closed'
  preferredContact: 'chat' | 'email'
  createdAt: string
  client?: { name: string } | null
}

const emit = defineEmits<{ select: [ticket: Ticket] }>()

const { data: ticketsData, pending: isLoading, refresh } = useFetch<Ticket[]>('/api/support/tickets')
const tickets = computed(() => ticketsData.value ?? [])

const searchQuery = ref('')
const statusFilter = ref('all')
const priorityFilter = ref('all')

const filtered = computed(() => {
  let items = tickets.value
  const q = searchQuery.value.trim().toLowerCase()
  if (q) {
    items = items.filter(t =>
      [t.ticketNumber, t.name, t.email, t.company, t.subject].filter(Boolean).some(v => String(v).toLowerCase().includes(q)),
    )
  }
  if (statusFilter.value !== 'all')
    items = items.filter(t => t.status === statusFilter.value)
  if (priorityFilter.value !== 'all')
    items = items.filter(t => t.priority === priorityFilter.value)
  return items
})

function statusVariant(status: string) {
  switch (status) {
    case 'open': return 'default'
    case 'in_progress': return 'secondary'
    case 'resolved': return 'outline'
    case 'closed': return 'secondary'
    default: return 'secondary'
  }
}

function priorityVariant(priority: string) {
  switch (priority) {
    case 'urgent': return 'destructive'
    case 'high': return 'destructive'
    case 'medium': return 'secondary'
    default: return 'outline'
  }
}

defineExpose({ refresh })
</script>

<template>
  <div class="space-y-4">
    <div class="flex flex-col md:flex-row gap-4 items-start md:items-center justify-between">
      <div class="relative flex-1 max-w-sm">
        <Search class="absolute left-2 top-2.5 h-4 w-4 text-muted-foreground" />
        <Input v-model="searchQuery" placeholder="Search tickets..." class="pl-8" />
      </div>
      <div class="flex gap-2 flex-wrap">
        <Select v-model="statusFilter">
          <SelectTrigger class="w-[160px]">
            <SelectValue placeholder="Status" />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value="all">
              All Status
            </SelectItem>
            <SelectItem value="open">
              Open
            </SelectItem>
            <SelectItem value="in_progress">
              In Progress
            </SelectItem>
            <SelectItem value="resolved">
              Resolved
            </SelectItem>
            <SelectItem value="closed">
              Closed
            </SelectItem>
          </SelectContent>
        </Select>
        <Select v-model="priorityFilter">
          <SelectTrigger class="w-[160px]">
            <SelectValue placeholder="Priority" />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value="all">
              All Priority
            </SelectItem>
            <SelectItem value="low">
              Low
            </SelectItem>
            <SelectItem value="medium">
              Medium
            </SelectItem>
            <SelectItem value="high">
              High
            </SelectItem>
            <SelectItem value="urgent">
              Urgent
            </SelectItem>
          </SelectContent>
        </Select>
      </div>
    </div>

    <div class="rounded-md border">
      <Table>
        <TableHeader>
          <TableRow>
            <TableHead>Ticket #</TableHead>
            <TableHead>Subject</TableHead>
            <TableHead>Requester</TableHead>
            <TableHead>Channel</TableHead>
            <TableHead>Priority</TableHead>
            <TableHead>Status</TableHead>
            <TableHead>Created</TableHead>
          </TableRow>
        </TableHeader>
        <TableBody v-if="isLoading">
          <TableRow v-for="i in 5" :key="`skeleton-${i}`">
            <TableCell colspan="7">
              <Skeleton class="h-6 w-full" />
            </TableCell>
          </TableRow>
        </TableBody>
        <TableBody v-else>
          <TableRow
            v-for="ticket in filtered"
            :key="ticket.id"
            class="cursor-pointer"
            @click="emit('select', ticket)"
          >
            <TableCell class="font-medium">
              {{ ticket.ticketNumber }}
            </TableCell>
            <TableCell class="max-w-xs truncate">
              {{ ticket.subject }}
            </TableCell>
            <TableCell>
              <div>{{ ticket.name }}</div>
              <div class="text-xs text-muted-foreground">
                {{ ticket.email }}
              </div>
            </TableCell>
            <TableCell>
              <Badge variant="outline" class="gap-1">
                <MessageCircle v-if="ticket.preferredContact === 'chat'" class="h-3 w-3" />
                <Mail v-else class="h-3 w-3" />
                {{ ticket.preferredContact === 'chat' ? 'Chat' : 'Email' }}
              </Badge>
            </TableCell>
            <TableCell>
              <Badge :variant="priorityVariant(ticket.priority)">
                {{ ticket.priority }}
              </Badge>
            </TableCell>
            <TableCell>
              <Badge :variant="statusVariant(ticket.status)">
                {{ ticket.status.replace('_', ' ') }}
              </Badge>
            </TableCell>
            <TableCell class="text-sm text-muted-foreground">
              {{ new Date(ticket.createdAt).toLocaleDateString() }}
            </TableCell>
          </TableRow>
          <TableRow v-if="filtered.length === 0">
            <TableCell colspan="7" class="text-center text-muted-foreground">
              No support tickets yet.
            </TableCell>
          </TableRow>
        </TableBody>
      </Table>
    </div>
  </div>
</template>
