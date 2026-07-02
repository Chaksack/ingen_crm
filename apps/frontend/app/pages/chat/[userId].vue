<script setup lang="ts">
interface CollabUser { id: string; display_name: string; email: string }
interface ChatMessage { id: string; sender_user_id: string; recipient_user_id: string; body: string; created_at: string }

const route = useRoute()
const peerId = computed(() => route.params.userId as string)

const api = useApi()
const auth = useAuthStore()
const chat = useChatStore()

const peer = ref<CollabUser | null>(null)
const draft = ref('')
const scrollEl = ref<HTMLElement | null>(null)
const breadcrumbOverride = useBreadcrumbOverride()

const messages = computed(() => chat.messagesByPeer[peerId.value] ?? [])

async function load() {
  const users = await api<CollabUser[]>('/collab/users')
  peer.value = users.find((u) => u.id === peerId.value) ?? null
  breadcrumbOverride.value = peer.value?.display_name ?? null
  const history = await api<ChatMessage[]>(`/collab/messages/${peerId.value}`)
  chat.setHistory(peerId.value, history ?? [])
  await nextTick()
  scrollEl.value?.scrollTo({ top: scrollEl.value.scrollHeight })
}

watch(messages, async () => {
  await nextTick()
  scrollEl.value?.scrollTo({ top: scrollEl.value.scrollHeight, behavior: 'smooth' })
})

watch(peerId, load, { immediate: true })

function sendMessage() {
  if (!draft.value.trim()) return
  chat.send(peerId.value, draft.value)
  draft.value = ''
}
</script>

<template>
  <div class="flex h-[calc(100vh-6.5rem)] flex-col">
    <h1 class="text-2xl font-semibold tracking-tight">{{ peer?.display_name ?? 'Chat' }}</h1>
    <p class="text-sm text-muted-foreground">{{ peer?.email }}</p>

    <div ref="scrollEl" class="mt-4 flex-1 space-y-2 overflow-y-auto rounded-md border border-border p-4">
      <div
        v-for="m in messages"
        :key="m.id"
        class="max-w-[70%] rounded-lg px-3 py-2 text-sm"
        :class="m.sender_user_id === auth.user?.id
          ? 'ml-auto bg-primary text-primary-foreground'
          : 'bg-muted'"
      >
        {{ m.body }}
      </div>
      <p v-if="messages.length === 0" class="text-sm text-muted-foreground">No messages yet. Say hello!</p>
    </div>

    <form class="mt-4 flex gap-2" @submit.prevent="sendMessage">
      <Input v-model="draft" placeholder="Type a message…" />
      <Button type="submit" :disabled="!chat.connected">Send</Button>
    </form>
    <p v-if="!chat.connected" class="mt-1 text-xs text-destructive">Reconnecting…</p>
  </div>
</template>
