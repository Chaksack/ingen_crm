<script setup lang="ts">
import { computed, nextTick, onMounted, ref, watch } from 'vue'

interface User {
  id: string
  name: string
  avatar: string
}

interface Message {
  id: string
  userId: string
  text: string
  createdAt: number
  attachments?: Attachment[]
}

interface Channel {
  id: string
  name: string
  type: 'channel' | 'dm'
  members?: string[]
}

interface Attachment {
  id: string
  name: string
  type: string
  size: number
  isImage: boolean
  url?: string // data URL for images or small files to persist preview
}

const currentUser: User = {
  id: 'u_self',
  name: 'Andrew Chakdahah',
  avatar: '/avatars/avatartion.png',
}

const users: Record<string, User> = {
  [currentUser.id]: currentUser,
  u_1: { id: 'u_1', name: 'IBS Support', avatar: '/avatars/03.png' },
  u_2: { id: 'u_2', name: 'Accounting', avatar: '/avatars/02.png' },
}

const channels = ref<Channel[]>([
  { id: 'c_general', name: 'general', type: 'channel', members: [currentUser.id, 'u_1', 'u_2'] },
  { id: 'c_ops', name: 'ops', type: 'channel', members: [currentUser.id, 'u_2'] },
  { id: 'dm_u_1', name: 'IBS Support', type: 'dm', members: [currentUser.id, 'u_1'] },
])

const activeId = ref<string>('c_general')
const draft = ref<string>('')
const messages = ref<Record<string, Message[]>>({})
const composerAttachments = ref<Attachment[]>([])
const composerInput = ref<HTMLTextAreaElement | null>(null)
const showEmoji = ref(false)
const fileInput = ref<HTMLInputElement | null>(null)

const storageKey = 'ibs_erp_chat_v1'

// Dialog and form state
const openDmDialog = ref(false)
const openNewChannel = ref(false)
const openAddPeople = ref(false)
const newChannelName = ref('')
const selectedUsersForChannel = ref<string[]>([])
const selectedUsersToAdd = ref<string[]>([])
const openCommand = ref(false)
const dmSearch = ref('')

// Layout controls to mimic Email layout (resizable + collapsible left nav)
const isCollapsed = ref(false)
const navCollapsedSize = 4
const defaultLayout = [18, 82]

function onCollapse() {
  isCollapsed.value = true
}

function onExpand() {
  isCollapsed.value = false
}

function loadState() {
  try {
    const raw = localStorage.getItem(storageKey)
    if (raw) {
      const data = JSON.parse(raw)
      if (data.channels && Array.isArray(data.channels)) {
        channels.value = data.channels
      }
      if (data.messages) {
        messages.value = data.messages
      }
      if (data.activeId) {
        activeId.value = data.activeId
      }
    }
  }
  catch {}
}

function saveState() {
  try {
    localStorage.setItem(storageKey, JSON.stringify({
      channels: channels.value,
      messages: messages.value,
      activeId: activeId.value,
    }))
  }
  catch {}
}

onMounted(() => {
  loadState()
  // migrate old default names for dm if missing members
  channels.value = channels.value.map((c) => {
    if (c.type === 'dm' && (!c.members || c.members.length < 2)) {
      const other = c.id.replace('dm_', '')
      return { ...c, members: [currentUser.id, other] }
    }
    return c
  })
  // Seed with a welcome message if empty
  if (!messages.value.c_general || messages.value.c_general.length === 0) {
    messages.value.c_general = [
      {
        id: crypto.randomUUID(),
        userId: 'u_1',
        text: 'Welcome to #general! This is the start of your Slack-like chat. 🎉',
        createdAt: Date.now() - 1000 * 60 * 5,
      },
    ]
  }
  saveState()
  scrollToBottom()
})

watch([messages, activeId, channels], () => {
  saveState()
  nextTick(() => scrollToBottom())
}, { deep: true })

const activeChannel = computed(() => channels.value.find(c => c.id === activeId.value))
const activeMessages = computed<Message[]>(() => messages.value[activeId.value] || [])

function dmDisplay(c: Channel) {
  const otherId = c.members?.find(u => u !== currentUser.id)
  const other = otherId ? users[otherId] : undefined
  return other?.name || c.name
}

function dmAvatar(c: Channel) {
  const otherId = c.members?.find(u => u !== currentUser.id)
  const other = otherId ? users[otherId] : undefined
  return other?.avatar || '/avatars/01.png'
}

function send() {
  const text = draft.value.trim()
  // Allow sending attachments without text, like Slack
  if (!text && composerAttachments.value.length === 0) {
    return
  }
  const list = messages.value[activeId.value] || []
  const msg: Message = {
    id: crypto.randomUUID(),
    userId: currentUser.id,
    text,
    createdAt: Date.now(),
    attachments: composerAttachments.value.length ? [...composerAttachments.value] : undefined,
  }
  messages.value = { ...messages.value, [activeId.value]: [...list, msg] }
  draft.value = ''
  composerAttachments.value = []
  autoResize()
}

function handleKeydown(e: KeyboardEvent) {
  if (e.key === 'Enter' && !e.shiftKey) {
    e.preventDefault()
    send()
  }
}

const scroller = ref<HTMLElement | null>(null)
function scrollToBottom() {
  if (!scroller.value) {
    return
  }
  scroller.value.scrollTop = scroller.value.scrollHeight
}

function autoResize() {
  const el = composerInput.value
  if (!el) {
    return
  }
  el.style.height = 'auto'
  const max = 200 // px
  el.style.height = `${Math.min(el.scrollHeight, max)}px`
}

function insertEmoji(emoji: string) {
  const el = composerInput.value
  if (!el) {
    draft.value += emoji
    return
  }
  const start = el.selectionStart || 0
  const end = el.selectionEnd || 0
  const before = draft.value.slice(0, start)
  const after = draft.value.slice(end)
  draft.value = before + emoji + after
  nextTick(() => {
    const pos = start + emoji.length
    el.selectionStart = el.selectionEnd = pos
    el.focus()
    autoResize()
  })
}

async function onFilesSelected(e: Event) {
  const input = e.target as HTMLInputElement
  const files = input.files
  if (!files || files.length === 0) {
    return
  }
  const list: Attachment[] = []
  const readers: Promise<void>[] = []
  for (const file of Array.from(files)) {
    const att: Attachment = {
      id: crypto.randomUUID(),
      name: file.name,
      type: file.type,
      size: file.size,
      isImage: file.type.startsWith('image/'),
    }
    if (att.isImage && file.size <= 2 * 1024 * 1024) { // <=2MB store as data URL
      const p = new Promise<void>((resolve) => {
        const reader = new FileReader()
        reader.onload = () => {
          att.url = String(reader.result)
          resolve()
        }
        reader.onerror = () => resolve()
        reader.readAsDataURL(file)
      })
      readers.push(p)
    }
    list.push(att)
  }
  if (readers.length) {
    await Promise.all(readers)
  }
  composerAttachments.value = [...composerAttachments.value, ...list]
  // reset input to allow re-uploading the same file name
  if (fileInput.value) {
    fileInput.value.value = ''
  }
}

function removeAttachment(id: string) {
  composerAttachments.value = composerAttachments.value.filter(a => a.id !== id)
}

function formatTime(ts: number) {
  const d = new Date(ts)
  return d.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })
}

function createOrOpenDm(userId: string) {
  if (!userId || userId === currentUser.id) {
    return
  }
  const id = `dm_${userId}`
  let dm = channels.value.find(c => c.id === id)
  if (!dm) {
    dm = { id, name: users[userId]?.name || 'Direct message', type: 'dm', members: [currentUser.id, userId] }
    channels.value = [dm, ...channels.value]
  }
  activeId.value = id
  openDmDialog.value = false
  openCommand.value = false
  dmSearch.value = ''
}

function toggleSelection(arrRef: typeof selectedUsersForChannel, id: string) {
  const set = new Set(arrRef.value)
  if (set.has(id)) {
    set.delete(id)
  }
  else {
    set.add(id)
  }
  arrRef.value = Array.from(set)
}

function createChannel() {
  const name = newChannelName.value.trim()
  if (!name) {
    return
  }
  const id = `c_${name.replace(/\s+/g, '-').toLowerCase()}`
  if (!channels.value.find(c => c.id === id)) {
    const members = Array.from(new Set([currentUser.id, ...selectedUsersForChannel.value]))
    const ch: Channel = { id, name, type: 'channel', members }
    channels.value = [ch, ...channels.value]
  }
  activeId.value = id
  newChannelName.value = ''
  selectedUsersForChannel.value = []
  openNewChannel.value = false
}

function addPeopleToChannel() {
  const ch = activeChannel.value
  if (!ch || ch.type !== 'channel') {
    return
  }
  const existing = new Set(ch.members || [])
  selectedUsersToAdd.value.forEach(u => existing.add(u))
  const updated = { ...ch, members: Array.from(existing) }
  channels.value = channels.value.map(c => c.id === ch.id ? updated : c)
  selectedUsersToAdd.value = []
  openAddPeople.value = false
}
</script>

<template>
  <div class="-m-4 lg:-m-6">
    <ResizablePanelGroup id="chat-resize-panel-group" direction="horizontal" class="h-[calc(100dvh-54px-3rem)] items-stretch">
      <ResizablePanel
        id="chat-left-panel"
        :default-size="defaultLayout[0]"
        :collapsed-size="navCollapsedSize"
        collapsible
        :min-size="15"
        :max-size="22"
        @expand="onExpand"
        @collapse="onCollapse"
      >
        <!-- Left rail structured like Mail layout: header / content / footer -->
        <div class="flex h-full min-h-0 flex-col">
          <!-- Sidebar Header -->
          <div :class="isCollapsed ? 'h-[56px]' : 'flex h-[56px] items-center px-2'">
            <div v-if="!isCollapsed" class="flex w-full items-center justify-between">
              <div class="flex items-center gap-2">
                <Icon name="i-lucide-message-square" />
                <span class="font-semibold">Chat</span>
              </div>
              <div class="flex items-center gap-1">
                <Button size="icon" variant="ghost" class="size-6" title="New channel" @click="openNewChannel = true">
                  <Icon name="i-lucide-hash" />
                </Button>
                <Button size="icon" variant="ghost" class="size-6" title="New DM" @click="openDmDialog = true">
                  <Icon name="i-lucide-user-plus" />
                </Button>
              </div>
            </div>
          </div>
          <Separator />
          <!-- Sidebar Content (scrollable) -->
          <div class="flex-1 overflow-y-auto">
            <div class="p-3" :class="isCollapsed ? 'px-1' : ''">
              <div v-if="!isCollapsed" class="flex items-center justify-between mb-2">
                <p class="text-xs uppercase text-muted-foreground">Channels</p>
                <Button size="icon" variant="ghost" class="size-6" @click="openNewChannel = true">
                  <Icon name="i-lucide-plus" />
                </Button>
              </div>
              <ul class="space-y-1">
                <li v-for="c in channels.filter(c => c.type === 'channel')" :key="c.id">
                  <button class="w-full text-left px-2 py-1 rounded hover:bg-muted flex items-center gap-2" :class="{ 'bg-muted': c.id === activeId }" @click="activeId = c.id">
                    <Icon name="i-lucide-hash" class="text-muted-foreground" />
                    <span v-if="!isCollapsed">#{{ c.name }}</span>
                  </button>
                </li>
              </ul>
            </div>
            <div class="p-3" :class="isCollapsed ? 'px-1' : ''">
              <div v-if="!isCollapsed" class="flex items-center justify-between mb-2">
                <p class="text-xs uppercase text-muted-foreground">Direct messages</p>
                <Button size="icon" variant="ghost" class="size-6" @click="openDmDialog = true">
                  <Icon name="i-lucide-plus" />
                </Button>
              </div>
              <ul class="space-y-1">
                <li v-for="c in channels.filter(c => c.type === 'dm')" :key="c.id">
                  <button class="w-full text-left px-2 py-1 rounded hover:bg-muted flex items-center gap-2" :class="{ 'bg-muted': c.id === activeId }" @click="activeId = c.id">
                    <Avatar class="size-5">
                      <AvatarImage :src="dmAvatar(c)" alt="DM" />
                      <AvatarFallback>DM</AvatarFallback>
                    </Avatar>
                    <span v-if="!isCollapsed">{{ dmDisplay(c) }}</span>
                  </button>
                </li>
              </ul>
            </div>
          </div>
          <Separator />
          <!-- Sidebar Footer -->
          <div class="px-3 py-2 text-xs text-muted-foreground" :class="isCollapsed ? 'px-1 text-center' : ''">
            <span v-if="!isCollapsed">You are signed in as {{ currentUser.name }}</span>
            <span v-else>•</span>
          </div>
        </div>
      </ResizablePanel>
      <ResizableHandle id="chat-resize-handle" with-handle />
      <ResizablePanel id="chat-right-panel" :default-size="defaultLayout[1]" :min-size="30">
        <div class="flex h-full min-h-0 flex-col overflow-hidden">
          <div class="flex h-[56px] items-center gap-2 border-b px-4 justify-between">
            <div class="flex items-center gap-2 min-w-0">
              <Icon name="i-lucide-hash" v-if="activeChannel?.type === 'channel'" />
              <Icon name="i-lucide-message-circle" v-else />
              <h2 class="text-xl font-semibold truncate">{{ activeChannel?.type === 'channel' ? `#${activeChannel?.name}` : (activeChannel ? dmDisplay(activeChannel) : '') }}</h2>
            </div>
            <div class="flex items-center gap-2">
              <Button v-if="activeChannel?.type === 'channel'" variant="ghost" size="sm" @click="openAddPeople = true">
                <Icon name="i-lucide-user-plus" class="mr-1" /> Add people
              </Button>
            </div>
          </div>
          <!-- Main chat area -->
          <section class="flex min-h-0 flex-1 flex-col overflow-hidden">
            <div ref="scroller" class="flex-1 overflow-y-auto px-4 py-4 pb-28 space-y-4">
              <div v-for="m in activeMessages" :key="m.id" class="flex items-start gap-3">
                <Avatar class="size-8">
                  <AvatarImage :src="users[m.userId]?.avatar || '/avatars/01.png'" :alt="users[m.userId]?.name || 'User'" />
                  <AvatarFallback>{{ users[m.userId]?.name?.[0] || 'U' }}</AvatarFallback>
                </Avatar>
                <div class="min-w-0">
                  <div class="flex items-baseline gap-2">
                    <span class="font-medium">{{ users[m.userId]?.name || 'Unknown' }}</span>
                    <span class="text-xs text-muted-foreground">{{ formatTime(m.createdAt) }}</span>
                  </div>
                  <p v-if="m.text" class="whitespace-pre-wrap break-words">{{ m.text }}</p>
                  <div v-if="m.attachments && m.attachments.length" class="mt-2 flex flex-wrap gap-2">
                    <template v-for="att in m.attachments" :key="att.id">
                      <div v-if="att.isImage && att.url" class="overflow-hidden rounded border bg-muted/20">
                        <img :src="att.url" :alt="att.name" class="max-h-40 object-contain" />
                      </div>
                      <div v-else class="flex items-center gap-2 rounded border px-2 py-1 text-sm">
                        <Icon name="i-lucide-paperclip" class="text-muted-foreground" />
                        <span class="truncate max-w-[200px]">{{ att.name }}</span>
                        <span class="text-xs text-muted-foreground">{{ Math.ceil(att.size / 1024) }} KB</span>
                      </div>
                    </template>
                  </div>
                </div>
              </div>
            </div>

            <!-- Fixed composer at the bottom of the messaging column -->
            <div class="z-10 border-t bg-background p-3 shrink-0 sticky bottom-0">
              <!-- Slack-like composer container -->
              <div class="rounded-lg border bg-background px-2 py-1 shadow-sm">
                <!-- attachment previews -->
                <div v-if="composerAttachments.length" class="mb-2 flex flex-wrap gap-2">
                  <div v-for="att in composerAttachments" :key="att.id" class="group relative rounded border">
                    <img v-if="att.isImage && att.url" :src="att.url" :alt="att.name" class="max-h-24 object-contain rounded" />
                    <div v-else class="flex items-center gap-2 px-2 py-1 text-sm">
                      <Icon name="i-lucide-paperclip" class="text-muted-foreground" />
                      <span class="truncate max-w-[160px]">{{ att.name }}</span>
                    </div>
                    <button class="absolute right-1 top-1 hidden rounded bg-black/60 p-1 text-white group-hover:block" @click="removeAttachment(att.id)" aria-label="Remove attachment">
                      <Icon name="i-lucide-x" class="size-3" />
                    </button>
                  </div>
                </div>
                <div class="flex items-end gap-2">
                  <Button variant="ghost" size="icon" class="mb-1" @click="fileInput?.click()" title="Add file">
                    <Icon name="i-lucide-plus" />
                  </Button>
                  <Textarea
                    ref="composerInput"
                    v-model="draft"
                    class="max-h-[200px] w-full resize-none border-0 p-0 focus-visible:ring-0"
                    placeholder="Message #{{ activeChannel?.name }}"
                    @keydown="handleKeydown"
                    @input="autoResize"
                    rows="1"
                  />
                  <div class="relative mb-1">
                    <Button variant="ghost" size="icon" @click="showEmoji = !showEmoji" title="Emoji">
                      <Icon name="i-lucide-smile" />
                    </Button>
                    <div v-if="showEmoji" class="absolute bottom-10 right-0 z-20 w-60 rounded-md border bg-popover p-2 shadow-md">
                      <div class="grid grid-cols-8 gap-1 text-xl">
                        <button v-for="e in ['😀', '😁', '😂', '😊', '😍', '😎', '🤔', '🙌', '👍', '🎉', '🔥', '✨', '💯', '🙏', '😅', '😉', '🥳', '😇', '🤖', '🧠', '📝', '📎', '✅', '❗']" :key="e" class="rounded hover:bg-muted" @click="insertEmoji(e); showEmoji = false">{{ e }}</button>
                      </div>
                    </div>
                  </div>
                  <Button :disabled="!draft.trim() && composerAttachments.length === 0" @click="send" class="mb-1">
                    <Icon name="i-lucide-send" class="mr-1" /> Send
                  </Button>
                  <input ref="fileInput" type="file" class="hidden" multiple @change="onFilesSelected" />
                </div>
              </div>
              <p class="text-[11px] text-muted-foreground mt-1">Enter to send • Shift+Enter for new line • Use + to add files</p>
            </div>
          </section>
          <!-- Panel Footer (separate from the sticky composer) -->
          <Separator />
          <div class="px-4 py-2 text-xs text-muted-foreground">
            Tip: Use / to open commands • Ctrl/Cmd+K to search
          </div>
        </div>
      </ResizablePanel>
    </ResizablePanelGroup>

    <!-- Start DM Dialog using Command palette style -->
    <Dialog v-model:open="openDmDialog">
      <DialogContent class="sm:max-w-[520px] p-0 overflow-hidden">
        <DialogHeader class="px-4 pt-4">
          <DialogTitle>Start a direct message</DialogTitle>
          <DialogDescription>Select a user to start a conversation.</DialogDescription>
        </DialogHeader>
        <div class="p-4 pt-0">
          <Command class="rounded-lg border shadow-none">
            <CommandInput v-model="dmSearch" placeholder="Search users..." />
            <CommandList>
              <CommandEmpty>No users found.</CommandEmpty>
              <CommandGroup heading="Users">
                <CommandItem v-for="u in Object.values(users).filter(u => u.id !== currentUser.id && (!dmSearch || u.name.toLowerCase().includes(dmSearch.toLowerCase())))" :key="u.id" @select="createOrOpenDm(u.id)">
                  <Avatar class="size-5 mr-2">
                    <AvatarImage :src="u.avatar" :alt="u.name" />
                    <AvatarFallback>{{ u.name[0] }}</AvatarFallback>
                  </Avatar>
                  <span class="truncate">{{ u.name }}</span>
                </CommandItem>
              </CommandGroup>
            </CommandList>
          </Command>
        </div>
        <DialogFooter class="px-4 pb-4">
          <DialogClose as-child>
            <Button variant="outline">Close</Button>
          </DialogClose>
        </DialogFooter>
      </DialogContent>
    </Dialog>

    <!-- New Channel Dialog -->
    <Dialog v-model:open="openNewChannel">
      <DialogContent class="sm:max-w-[560px]">
        <DialogHeader>
          <DialogTitle>Create a channel</DialogTitle>
          <DialogDescription>Name your channel and add members.</DialogDescription>
        </DialogHeader>
        <div class="space-y-4">
          <div class="space-y-2">
            <Label for="channel_name">Channel name</Label>
            <Input id="channel_name" v-model="newChannelName" placeholder="e.g. marketing" />
          </div>
          <div class="space-y-2">
            <Label>Add members</Label>
            <div class="max-h-48 overflow-auto border rounded-md p-2 space-y-1">
              <label v-for="u in Object.values(users).filter(u => u.id !== currentUser.id)" :key="u.id" class="flex items-center gap-2 px-1 py-1 rounded hover:bg-muted cursor-pointer">
                <Checkbox :checked="selectedUsersForChannel.includes(u.id)" @update:checked="() => toggleSelection(selectedUsersForChannel, u.id)" />
                <Avatar class="size-5">
                  <AvatarImage :src="u.avatar" :alt="u.name" />
                  <AvatarFallback>{{ u.name[0] }}</AvatarFallback>
                </Avatar>
                <span class="truncate">{{ u.name }}</span>
              </label>
            </div>
          </div>
        </div>
        <DialogFooter>
          <DialogClose as-child>
            <Button variant="outline">Cancel</Button>
          </DialogClose>
          <Button :disabled="!newChannelName.trim()" @click="createChannel">Create</Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>

    <!-- Add People Dialog -->
    <Dialog v-model:open="openAddPeople">
      <DialogContent class="sm:max-w-[520px]">
        <DialogHeader>
          <DialogTitle>Add people to #{{ activeChannel?.name }}</DialogTitle>
          <DialogDescription>Select users to add to this channel.</DialogDescription>
        </DialogHeader>
        <div class="max-h-64 overflow-auto border rounded-md p-2 space-y-1">
          <label v-for="u in Object.values(users).filter(u => u.id !== currentUser.id)" :key="u.id" class="flex items-center gap-2 px-1 py-1 rounded hover:bg-muted cursor-pointer">
            <Checkbox :checked="selectedUsersToAdd.includes(u.id)" @update:checked="() => toggleSelection(selectedUsersToAdd, u.id)" />
            <Avatar class="size-5">
              <AvatarImage :src="u.avatar" :alt="u.name" />
              <AvatarFallback>{{ u.name[0] }}</AvatarFallback>
            </Avatar>
            <span class="truncate">{{ u.name }}</span>
          </label>
        </div>
        <DialogFooter>
          <DialogClose as-child>
            <Button variant="outline">Cancel</Button>
          </DialogClose>
          <Button @click="addPeopleToChannel">Add</Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  </div>
</template>

<style scoped>
/* no additional styles; rely on design system classes */
</style>
