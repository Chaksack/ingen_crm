<script lang="ts" setup>
import type { Mail } from './data/mails'
import type { LinkProp } from '~/components/mail/Nav.vue'
import { useMediaQuery } from '@vueuse/core'
import { ListFilter, PenSquare, Search } from 'lucide-vue-next'
import { cn } from '~/lib/utils'

const props = withDefaults(defineProps<MailProps>(), {
  defaultCollapsed: false,
  defaultLayout: () => [18, 32, 50],
})

interface MailProps {
  accounts: {
    label: string
    email: string
    icon: string
  }[]
  mails: Mail[]
  defaultLayout?: number[]
  defaultCollapsed?: boolean
  navCollapsedSize: number
}

const isCollapsed = ref(props.defaultCollapsed)
const selectedMail = ref<string | undefined>()
const searchValue = ref('')
const debouncedSearch = refDebounced(searchValue, 250)
const activeTab = ref<'all' | 'unread'>('all')

const filteredMailList = computed(() => {
  let output: Mail[]
  const searchValue = debouncedSearch.value?.trim()
  if (!searchValue) {
    output = props.mails
  }
  else {
    output = props.mails.filter((item) => {
      return item.name.includes(debouncedSearch.value)
        || item.email.includes(debouncedSearch.value)
        || item.subject.includes(debouncedSearch.value)
        || item.text.includes(debouncedSearch.value)
    })
  }

  return activeTab.value === 'unread' ? output.filter(item => !item.read) : output
})

const selectedMailData = computed(() => props.mails.find(item => item.id === selectedMail.value))

const links: LinkProp[] = [
  {
    title: 'Inbox',
    label: '128',
    icon: 'lucide:inbox',
    variant: 'default',
  },
  {
    title: 'Drafts',
    label: '9',
    icon: 'lucide:file',
    variant: 'ghost',
  },
  {
    title: 'Sent Items',
    label: '',
    icon: 'lucide:send',
    variant: 'ghost',
  },
  {
    title: 'Junk Email',
    label: '23',
    icon: 'lucide:shield-alert',
    variant: 'ghost',
  },
  {
    title: 'Deleted Items',
    label: '',
    icon: 'lucide:trash',
    variant: 'ghost',
  },
  {
    title: 'Archive',
    label: '',
    icon: 'lucide:archive',
    variant: 'ghost',
  },
]

const links2: LinkProp[] = [
  {
    title: 'Social',
    label: '972',
    icon: 'lucide:user-2',
    variant: 'ghost',
  },
  {
    title: 'Updates',
    label: '342',
    icon: 'lucide:alert-circle',
    variant: 'ghost',
  },
  {
    title: 'Forums',
    label: '128',
    icon: 'lucide:message-square',
    variant: 'ghost',
  },
  {
    title: 'Shopping',
    label: '8',
    icon: 'lucide:shopping-cart',
    variant: 'ghost',
  },
  {
    title: 'Promotions',
    label: '21',
    icon: 'lucide:tag',
    variant: 'ghost',
  },
]

function onCollapse() {
  isCollapsed.value = true
}

function onExpand() {
  isCollapsed.value = false
}

const defaultCollapse = useMediaQuery('(max-width: 768px)')

watch(() => defaultCollapse.value, () => {
  isCollapsed.value = defaultCollapse.value
})
</script>

<template>
  <TooltipProvider :delay-duration="0">
    <ResizablePanelGroup
      id="resize-panel-group-1"
      direction="horizontal"
      class="h-full max-h-[calc(100dvh-54px-3rem)] items-stretch rounded-lg border"
    >
      <ResizablePanel
        id="resize-panel-1"
        :default-size="defaultLayout[0]"
        :collapsed-size="navCollapsedSize"
        collapsible
        :min-size="15"
        :max-size="22"
        :class="cn('flex h-full flex-col bg-muted/30', isCollapsed && 'min-w-[50px] transition-all duration-300 ease-in-out')"
        @expand="onExpand"
        @collapse="onCollapse"
      >
        <div :class="cn('flex h-[52px] shrink-0 items-center', isCollapsed ? 'justify-center' : 'px-2.5')">
          <MailAccountSwitcher :is-collapsed="isCollapsed" :accounts="accounts" />
        </div>

        <div :class="cn('shrink-0 pb-2', isCollapsed ? 'flex justify-center' : 'px-2.5')">
          <Tooltip v-if="isCollapsed" :delay-duration="0">
            <TooltipTrigger as-child>
              <Button size="icon" class="size-9 rounded-full bg-[#0078D4] text-white hover:bg-[#106EBE]">
                <PenSquare class="size-4" />
                <span class="sr-only">New message</span>
              </Button>
            </TooltipTrigger>
            <TooltipContent side="right">
              New message
            </TooltipContent>
          </Tooltip>
          <Button v-else class="w-full justify-center gap-2 rounded-md bg-[#0078D4] font-medium text-white hover:bg-[#106EBE]">
            <PenSquare class="size-4" />
            New message
          </Button>
        </div>

        <Separator />
        <div class="min-h-0 flex-1 overflow-y-auto">
          <MailNav
            :is-collapsed="isCollapsed"
            :links="links"
          />
          <Separator class="my-1" />
          <MailNav
            :is-collapsed="isCollapsed"
            :links="links2"
          />
        </div>
      </ResizablePanel>
      <ResizableHandle id="resize-handle-1" with-handle />
      <ResizablePanel id="resize-panel-2" :default-size="defaultLayout[1]" :min-size="25" class="flex h-full flex-col">
        <div class="flex h-[52px] shrink-0 items-center gap-4 border-b px-4">
          <button
            type="button"
            :class="cn(
              'border-b-2 py-2 text-sm font-medium transition-colors -mb-px',
              activeTab === 'all' ? 'border-[#0078D4] text-[#0078D4] dark:text-[#4CC2FF]' : 'border-transparent text-muted-foreground hover:text-foreground',
            )"
            @click="activeTab = 'all'"
          >
            All mail
          </button>
          <button
            type="button"
            :class="cn(
              'border-b-2 py-2 text-sm font-medium transition-colors -mb-px',
              activeTab === 'unread' ? 'border-[#0078D4] text-[#0078D4] dark:text-[#4CC2FF]' : 'border-transparent text-muted-foreground hover:text-foreground',
            )"
            @click="activeTab = 'unread'"
          >
            Unread
          </button>
          <ListFilter class="ml-auto size-4 text-muted-foreground" />
        </div>
        <div class="shrink-0 p-3">
          <div class="relative">
            <Search class="absolute left-2.5 top-2.5 size-4 text-muted-foreground" />
            <Input v-model="searchValue" placeholder="Search mail" class="rounded-md bg-background pl-8" />
          </div>
        </div>
        <MailList v-model:selected-mail="selectedMail" :items="filteredMailList" />
      </ResizablePanel>
      <ResizableHandle id="resize-handle-2" with-handle />
      <ResizablePanel id="resize-panel-3" :default-size="defaultLayout[2]" :min-size="30">
        <MailDisplay :mail="selectedMailData" @close="selectedMail = undefined" />
      </ResizablePanel>
    </ResizablePanelGroup>
  </TooltipProvider>
</template>
