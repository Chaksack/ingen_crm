<script setup lang="ts">
import { Loader2 } from 'lucide-vue-next'
import { toast } from 'vue-sonner'

interface Message {
  id: string
  authorType: 'staff' | 'client'
  authorName: string
  body: string
  createdAt: string
}

interface TicketDetail {
  id: string
  ticketNumber: string
  name: string
  email: string
  phone?: string
  company?: string
  subject: string
  description: string
  category?: string
  priority: 'low' | 'medium' | 'high' | 'urgent'
  status: 'open' | 'in_progress' | 'resolved' | 'closed'
  preferredContact: 'chat' | 'email'
  createdAt: string
  messages: Message[]
}

const props = defineProps<{ ticketId: string | null, open: boolean }>()
const emit = defineEmits<{ 'update:open': [value: boolean], 'changed': [] }>()

const ticket = ref<TicketDetail | null>(null)
const isLoading = ref(false)
const isSaving = ref(false)
const isSending = ref(false)
const replyBody = ref('')
let pollHandle: ReturnType<typeof setInterval> | undefined

async function loadTicket({ silent = false } = {}) {
  if (!props.ticketId)
    return
  if (!silent)
    isLoading.value = true
  try {
    ticket.value = await $fetch<TicketDetail>(`/api/support/tickets/${props.ticketId}`)
  }
  catch {
    if (!silent)
      toast.error('Failed to load ticket')
  }
  finally {
    isLoading.value = false
  }
}

watch(() => [props.ticketId, props.open], ([id, open]) => {
  clearInterval(pollHandle)
  if (id && open) {
    loadTicket()
    // Poll while the sheet is open so a customer's chat replies show up without
    // the staff member needing to close and reopen the ticket.
    pollHandle = setInterval(() => loadTicket({ silent: true }), 5000)
  }
}, { immediate: true })

onUnmounted(() => clearInterval(pollHandle))

async function updateField(patch: Record<string, unknown>) {
  if (!ticket.value)
    return
  isSaving.value = true
  try {
    await $fetch(`/api/support/tickets/${ticket.value.id}`, { method: 'PATCH', body: patch })
    Object.assign(ticket.value, patch)
    emit('changed')
  }
  catch {
    toast.error('Failed to update ticket')
  }
  finally {
    isSaving.value = false
  }
}

async function sendReply() {
  if (!ticket.value || !replyBody.value.trim())
    return
  isSending.value = true
  try {
    const message = await $fetch<Message>(`/api/support/tickets/${ticket.value.id}/messages`, {
      method: 'POST',
      body: { body: replyBody.value },
    })
    ticket.value.messages.push(message)
    if (ticket.value.status === 'open')
      ticket.value.status = 'in_progress'
    replyBody.value = ''
    emit('changed')
  }
  catch {
    toast.error('Failed to send reply')
  }
  finally {
    isSending.value = false
  }
}
</script>

<template>
  <Sheet :open="open" @update:open="(v) => emit('update:open', v)">
    <SheetContent side="right" class="w-full sm:max-w-xl overflow-y-auto">
      <SheetHeader>
        <div class="flex items-center gap-2">
          <SheetTitle>{{ ticket?.ticketNumber || 'Ticket' }}</SheetTitle>
          <Badge v-if="ticket" variant="outline" class="gap-1">
            <Icon :name="ticket.preferredContact === 'chat' ? 'i-lucide-message-circle' : 'i-lucide-mail'" class="h-3 w-3" />
            {{ ticket.preferredContact === 'chat' ? 'Chat' : 'Email' }}
          </Badge>
        </div>
        <SheetDescription>{{ ticket?.subject }}</SheetDescription>
      </SheetHeader>

      <div v-if="isLoading" class="px-4 py-6 text-sm text-muted-foreground">
        Loading ticket...
      </div>

      <div v-else-if="ticket" class="px-4 space-y-6">
        <div class="grid grid-cols-2 gap-4">
          <div class="grid gap-2">
            <Label>Status</Label>
            <Select :model-value="ticket.status" :disabled="isSaving" @update:model-value="(v) => updateField({ status: v })">
              <SelectTrigger>
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
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
          </div>
          <div class="grid gap-2">
            <Label>Priority</Label>
            <Select :model-value="ticket.priority" :disabled="isSaving" @update:model-value="(v) => updateField({ priority: v })">
              <SelectTrigger>
                <SelectValue />
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
                <SelectItem value="urgent">
                  Urgent
                </SelectItem>
              </SelectContent>
            </Select>
          </div>
        </div>

        <Card>
          <CardHeader>
            <h3 class="text-sm font-semibold">
              Requester
            </h3>
          </CardHeader>
          <CardContent class="space-y-2 text-sm">
            <div class="grid grid-cols-3 gap-2">
              <span class="text-muted-foreground">Name</span>
              <span class="col-span-2 font-medium">{{ ticket.name }}</span>
            </div>
            <div class="grid grid-cols-3 gap-2">
              <span class="text-muted-foreground">Email</span>
              <span class="col-span-2 font-medium">{{ ticket.email }}</span>
            </div>
            <div v-if="ticket.phone" class="grid grid-cols-3 gap-2">
              <span class="text-muted-foreground">Phone</span>
              <span class="col-span-2 font-medium">{{ ticket.phone }}</span>
            </div>
            <div v-if="ticket.company" class="grid grid-cols-3 gap-2">
              <span class="text-muted-foreground">Company</span>
              <span class="col-span-2 font-medium">{{ ticket.company }}</span>
            </div>
          </CardContent>
        </Card>

        <div>
          <h3 class="text-sm font-semibold mb-2">
            Description
          </h3>
          <p class="text-sm whitespace-pre-wrap text-muted-foreground">
            {{ ticket.description }}
          </p>
        </div>

        <div>
          <h3 class="text-sm font-semibold mb-2">
            Conversation
          </h3>
          <div v-if="ticket.messages.length === 0" class="text-sm text-muted-foreground">
            No replies yet.
          </div>
          <div v-else class="space-y-3">
            <div
              v-for="message in ticket.messages"
              :key="message.id"
              class="rounded-md border p-3 text-sm"
              :class="message.authorType === 'staff' ? 'bg-muted/50' : ''"
            >
              <div class="flex items-center justify-between mb-1">
                <span class="font-medium">{{ message.authorName }}</span>
                <span class="text-xs text-muted-foreground">{{ new Date(message.createdAt).toLocaleString() }}</span>
              </div>
              <p class="whitespace-pre-wrap">
                {{ message.body }}
              </p>
            </div>
          </div>

          <div class="mt-3 grid gap-2">
            <Textarea v-model="replyBody" rows="3" placeholder="Write a reply..." :disabled="isSending" />
            <Button size="sm" class="justify-self-end" :disabled="isSending || !replyBody.trim()" @click="sendReply">
              <Loader2 v-if="isSending" class="mr-2 h-4 w-4 animate-spin" />
              Send Reply
            </Button>
          </div>
        </div>
      </div>
    </SheetContent>
  </Sheet>
</template>
