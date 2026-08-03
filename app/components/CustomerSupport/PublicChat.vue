<script setup lang="ts">
import { Loader2, Send } from 'lucide-vue-next'

interface Message {
  id: string
  authorType: 'staff' | 'client'
  authorName: string
  body: string
  createdAt: string
}

interface TicketChat {
  id: string
  ticketNumber: string
  name: string
  subject: string
  status: 'open' | 'in_progress' | 'resolved' | 'closed'
  createdAt: string
  messages: Message[]
}

const props = defineProps<{ ticketId: string, accessToken: string, description: string }>()
const emit = defineEmits<{ endChat: [] }>()

const ticket = ref<TicketChat | null>(null)
const isLoading = ref(true)
const isSending = ref(false)
const draft = ref('')
const scrollEl = ref<HTMLElement | null>(null)
let pollHandle: ReturnType<typeof setInterval> | undefined

function scrollToBottom() {
  nextTick(() => {
    if (scrollEl.value)
      scrollEl.value.scrollTop = scrollEl.value.scrollHeight
  })
}

async function load({ silent = false } = {}) {
  if (!silent)
    isLoading.value = true
  try {
    const data = await $fetch<TicketChat>(`/api/support/tickets/public/${props.ticketId}`, {
      query: { token: props.accessToken },
    })
    const hadFewerMessages = (ticket.value?.messages.length ?? 0) < data.messages.length
    ticket.value = data
    if (hadFewerMessages)
      scrollToBottom()
  }
  finally {
    isLoading.value = false
  }
}

async function send() {
  if (!draft.value.trim())
    return
  const body = draft.value.trim()
  draft.value = ''
  isSending.value = true
  try {
    const message = await $fetch<Message>(`/api/support/tickets/public/${props.ticketId}/messages`, {
      method: 'POST',
      body: { token: props.accessToken, body },
    })
    ticket.value?.messages.push(message)
    scrollToBottom()
  }
  finally {
    isSending.value = false
  }
}

const statusLabel = computed(() => {
  switch (ticket.value?.status) {
    case 'open': return 'Waiting for a reply'
    case 'in_progress': return 'Agent is on this'
    case 'resolved': return 'Resolved'
    case 'closed': return 'Closed'
    default: return ''
  }
})

onMounted(async () => {
  await load()
  scrollToBottom()
  pollHandle = setInterval(() => load({ silent: true }), 4000)
})

onUnmounted(() => {
  if (pollHandle)
    clearInterval(pollHandle)
})
</script>

<template>
  <Card class="flex flex-col h-[70vh]">
    <CardHeader class="border-b">
      <div class="flex items-center justify-between gap-2">
        <div>
          <CardTitle class="text-base">
            {{ ticket?.subject || 'Support chat' }}
          </CardTitle>
          <CardDescription>{{ ticket?.ticketNumber }}</CardDescription>
        </div>
        <Badge v-if="ticket" variant="secondary">
          {{ statusLabel }}
        </Badge>
      </div>
    </CardHeader>

    <div v-if="isLoading" class="flex-1 flex items-center justify-center text-sm text-muted-foreground">
      <Loader2 class="mr-2 h-4 w-4 animate-spin" /> Connecting...
    </div>

    <template v-else>
      <div ref="scrollEl" class="flex-1 overflow-y-auto p-4 space-y-3">
        <div class="flex justify-end">
          <div class="max-w-[80%] rounded-lg bg-primary text-primary-foreground px-3 py-2 text-sm whitespace-pre-wrap">
            {{ description }}
          </div>
        </div>
        <div v-for="message in ticket?.messages" :key="message.id" class="flex" :class="message.authorType === 'client' ? 'justify-end' : 'justify-start'">
          <div
            class="max-w-[80%] rounded-lg px-3 py-2 text-sm whitespace-pre-wrap"
            :class="message.authorType === 'client' ? 'bg-primary text-primary-foreground' : 'bg-muted'"
          >
            <div v-if="message.authorType === 'staff'" class="text-xs font-medium text-muted-foreground mb-0.5">
              {{ message.authorName }}
            </div>
            {{ message.body }}
          </div>
        </div>
        <div v-if="ticket?.messages.length === 0" class="text-center text-xs text-muted-foreground pt-2">
          A support agent will join the conversation shortly.
        </div>
      </div>

      <div class="border-t p-3">
        <div v-if="ticket?.status === 'closed'" class="text-center text-sm text-muted-foreground py-2">
          This conversation is closed.
          <button type="button" class="underline" @click="emit('endChat')">
            Start a new one
          </button>
        </div>
        <form v-else class="flex items-end gap-2" @submit.prevent="send">
          <Textarea v-model="draft" rows="1" placeholder="Type a message..." class="min-h-9 resize-none" :disabled="isSending" @keydown.enter.exact.prevent="send" />
          <Button type="submit" size="icon" :disabled="isSending || !draft.trim()">
            <Loader2 v-if="isSending" class="h-4 w-4 animate-spin" />
            <Send v-else class="h-4 w-4" />
          </Button>
        </form>
      </div>
    </template>
  </Card>
</template>
