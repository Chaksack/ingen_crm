<script setup lang="ts">
interface Notification {
  id: string
  title: string
  body: string
  link: string | null
  is_read: boolean
  created_at: string
}

const api = useApi()
const notifications = ref<Notification[]>([])
const unreadCount = ref(0)
let poller: ReturnType<typeof setInterval> | undefined

const push = usePushNotifications()
async function togglePush() {
  if (push.subscribed.value) await push.unsubscribe()
  else await push.subscribe()
}

async function loadUnreadCount() {
  try {
    const res = await api<{ count: number }>('/notifications/unread-count')
    unreadCount.value = res?.count ?? 0
  } catch {
    // notification polling failures shouldn't surface as user-facing errors
  }
}

async function loadNotifications() {
  try {
    notifications.value = await api<Notification[]>('/notifications') ?? []
  } catch {
    notifications.value = []
  }
}

async function onOpenChange(open: boolean) {
  if (open) await loadNotifications()
}

async function markRead(n: Notification) {
  if (n.is_read) return
  n.is_read = true
  unreadCount.value = Math.max(0, unreadCount.value - 1)
  try {
    await api(`/notifications/${n.id}/read`, { method: 'POST' })
  } catch {
    // best-effort; the next poll reconciles state if this failed
  }
}

function formatTime(iso: string) {
  return new Date(iso).toLocaleString()
}

onMounted(() => {
  loadUnreadCount()
  poller = setInterval(loadUnreadCount, 30000)
})
onUnmounted(() => {
  if (poller) clearInterval(poller)
})
</script>

<template>
  <DropdownMenu @update:open="onOpenChange">
    <DropdownMenuTrigger as-child>
      <Button variant="ghost" size="icon" class="relative">
        <Icon name="lucide:bell" class="size-4.5" />
        <span
          v-if="unreadCount > 0"
          class="absolute -top-0.5 -right-0.5 flex h-4 min-w-4 items-center justify-center rounded-full bg-destructive px-1 text-[10px] font-medium text-destructive-foreground"
        >
          {{ unreadCount > 9 ? '9+' : unreadCount }}
        </span>
      </Button>
    </DropdownMenuTrigger>
    <DropdownMenuContent align="end" class="w-80">
      <DropdownMenuLabel>Notifications</DropdownMenuLabel>
      <DropdownMenuItem
        v-if="push.supported && push.permission.value !== 'denied'"
        class="text-sm text-muted-foreground"
        @select.prevent="togglePush"
      >
        <Icon :name="push.subscribed.value ? 'lucide:bell-off' : 'lucide:bell-plus'" class="mr-1.5 size-3.5" />
        {{ push.subscribed.value ? 'Disable push notifications' : 'Enable push notifications' }}
      </DropdownMenuItem>
      <p v-if="push.error.value" class="px-2 pb-1 text-xs text-destructive">{{ push.error.value }}</p>
      <DropdownMenuSeparator />
      <p v-if="notifications.length === 0" class="px-2 py-4 text-center text-sm text-muted-foreground">
        You're all caught up.
      </p>
      <div v-else class="max-h-96 overflow-y-auto">
        <DropdownMenuItem
          v-for="n in notifications"
          :key="n.id"
          class="flex flex-col items-start gap-0.5 whitespace-normal py-2"
          @select="markRead(n)"
        >
          <NuxtLink v-if="n.link" :to="n.link" class="w-full">
            <span class="flex items-center gap-1.5 font-medium">
              <span v-if="!n.is_read" class="size-1.5 rounded-full bg-primary" />
              {{ n.title }}
            </span>
            <span class="text-xs text-muted-foreground">{{ n.body }}</span>
            <span class="text-[10px] text-muted-foreground">{{ formatTime(n.created_at) }}</span>
          </NuxtLink>
          <template v-else>
            <span class="flex items-center gap-1.5 font-medium">
              <span v-if="!n.is_read" class="size-1.5 rounded-full bg-primary" />
              {{ n.title }}
            </span>
            <span class="text-xs text-muted-foreground">{{ n.body }}</span>
            <span class="text-[10px] text-muted-foreground">{{ formatTime(n.created_at) }}</span>
          </template>
        </DropdownMenuItem>
      </div>
    </DropdownMenuContent>
  </DropdownMenu>
</template>
