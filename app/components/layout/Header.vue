<script setup lang="ts">
import type { NotificationItems } from '~/types/notifications'
import { ref, watch } from 'vue'
import { useRoute } from 'vue-router'
import { notificationItems } from '~/constants/notifications'

function resolveNotificationItemsComponent(item: NotificationItems): any {
  if ('children' in item)
    return resolveComponent('LayoutSidebarNavGroup')

  return resolveComponent('LayoutSidebarNavLink')
}

const route = useRoute()

// ✅ Put messages OUTSIDE the function (global reactive state)
const messages = ref([
  {
    name: 'William Smith',
    email: 'williamsmith@example.com',
    time: '2 hrs ago',
    subject: 'Meeting Tomorrow',
    content: 'Hi, let\'s have a meeting tomorrow to discuss...',
    avatar: 'https://randomuser.me/api/portraits/men/75.jpg',
  },
  {
    name: 'Joseph Peters',
    email: 'josephpeters@example.com',
    time: '4 hrs ago',
    subject: 'Credit Enquiry Details',
    content: 'I\'ve been reviewing your credit enquiry...',
    avatar: 'https://randomuser.me/api/portraits/men/24.jpg',
  },
])

// EMPTY notifications for testing
const notifications = ref([])

function setLinks() {
  if (route.fullPath === '/') {
    return [{ title: 'Home', href: '/' }]
  }

  const segments = route.fullPath.split('/').filter(item => item !== '')

  const breadcrumbs = segments.map((item, index) => {
    const str = item.replace(/-/g, ' ')
    const title = str
      .split(' ')
      .map(word => word.charAt(0).toUpperCase() + word.slice(1).toLowerCase())
      .join(' ')

    return {
      title,
      href: `/${segments.slice(0, index + 1).join('/')}`,
    }
  })

  return [{ title: 'Home', href: '/' }, ...breadcrumbs]
}

const links = ref(setLinks())

watch(
  () => route.fullPath,
  () => {
    links.value = setLinks()
  },
)
</script>

<template>
  <header class="sticky top-0 md:peer-data-[variant=inset]:top-2 z-10 h-(--header-height) flex items-center gap-4 border-b bg-background px-4 md:px-6 md:rounded-tl-xl md:rounded-tr-xl">
    <div class="w-full flex items-center gap-4 h-4">
      <SidebarTrigger />
      <Separator orientation="vertical" />
      <BaseBreadcrumbCustom :links="links" />
      <div class="flex flex-col sm:flex-row ml-auto gap-2 md:gap-4 flex-wrap items-center">
        <div class="w-full sm:w-auto">
          <Search />
        </div>
        <Sheet>
          <SheetTrigger as-child>
            <Button variant="outline" class="h-8 px-2 md:px-4">
              <Icon name="i-lucide-loader-pinwheel" class="w-4 h-4" />
              <span class="hidden md:inline ml-2">Ask Ai</span>
            </Button>
          </SheetTrigger>
          <SheetContent class="px-2 overflow-y-auto rounded-l-lg ">
            <SheetHeader>
              <SheetTitle>Ask Anything</SheetTitle>
              <SheetDescription>
                Ask questions or get help with tasks using the AI assistant.
              </SheetDescription>
            </SheetHeader>
            <AiAssistant />
          </SheetContent>
          <SheetContent class="px-2 overflow-y-auto rounded-l-lg ">
            <SheetHeader>
              <SheetTitle>Ask Anything</SheetTitle>
              <SheetDescription>
                Ask questions or get help with tasks using the AI assistant.
              </SheetDescription>
            </SheetHeader>
            <AiAssistant />
          </SheetContent>
        </Sheet>
        <DropdownMenu>
          <DropdownMenuTrigger as-child>
            <div class="relative ml-auto">
              <Button variant="outline" size="icon" class="h-8 w-8">
                <Icon name="i-lucide-mail" class="w-4 h-4" />
              </Button>

              <!-- Badge properly positioned -->
              <Badge
                class="absolute -top-1 -right-1 h-4 min-w-4 flex items-center justify-center
               rounded-full px-1 text-[10px] font-mono tabular-nums"
              >
                8
              </Badge>
            </div>
          </DropdownMenuTrigger>

          <DropdownMenuContent side="top" class="mr-10 mt-2 rounded-lg">
            <!-- If EMAILS exist -->
            <template v-if="messages.length > 0">
              <template v-for="(msg, i) in messages" :key="i">
                <div class="text-sm items-center mx-2 mb-2 gap-2 md:flex">
                  <Avatar class="h-6 w-6 rounded-lg">
                    <AvatarImage :src="msg.avatar" alt="" />
                    <AvatarFallback>{{ msg.name[0] }}</AvatarFallback>
                  </Avatar>
                  {{ msg.name }}
                  <span class="text-muted-foreground text-xs"> {{ msg.email }} </span>
                  <span class="text-muted-foreground text-xs ml-auto"> {{ msg.time }} </span>
                </div>

                <div class="flex-1 justify-center">
                  <p class="text-xs mx-2 text-muted-foreground mb-2">
                    subject: {{ msg.subject }}
                  </p>
                  <p class="text-xs mx-2 text-muted-foreground mb-4">
                    {{ msg.content }}
                  </p>
                </div>

                <DropdownMenuSeparator v-if="i !== messages.length - 1" />
              </template>

              <DropdownMenuSeparator />

              <DropdownMenuItem>
                <NuxtLink to="/email" class="w-full text-center">
                  View All Messages
                </NuxtLink>
              </DropdownMenuItem>
            </template>

            <!-- If EMPTY -->
            <template v-else>
              <Empty>
                <EmptyHeader>
                  <EmptyMedia variant="icon">
                    <Icon name="i-lucide-folder-code" class="size-6" />
                  </EmptyMedia>
                  <EmptyTitle>No Messages</EmptyTitle>
                  <EmptyDescription>You have no new messages.</EmptyDescription>
                </EmptyHeader>
              </Empty>
            </template>
          </DropdownMenuContent>
        </DropdownMenu>

        <DropdownMenu>
          <DropdownMenuTrigger as-child>
            <div class="relative ml-auto">
              <Button variant="outline" size="icon" class="h-8 w-8">
                <Icon name="i-lucide-bell" class="w-4 h-4" />
              </Button>

              <!-- Badge properly positioned -->
              <Badge
                class="absolute -top-1 -right-1 h-4 min-w-4 flex items-center justify-center
               rounded-full px-1 text-[10px] font-mono tabular-nums"
              >
                8
              </Badge>
            </div>
          </DropdownMenuTrigger>
          <DropdownMenuContent side="top" class="mr-10 mt-2 rounded-lg">
            <!-- If NOTIFICATIONS exist -->
            <template v-if="notifications.length > 0">
              <template v-for="(item, i) in notifications" :key="i">
                <div class="text-sm items-center mx-2 mb-2 gap-2 md:flex">
                  <img
                    :src="item.icon"
                    height="24" width="24"
                    class="rounded-md object-cover"
                  >
                  {{ item.title }}
                </div>

                <p class="text-xs mx-2 text-muted-foreground mb-4">
                  {{ item.description }}
                </p>

                <DropdownMenuSeparator v-if="i !== notifications.length - 1" />
              </template>
            </template>

            <!-- If EMPTY -->
            <template v-else>
              <Empty>
                <EmptyHeader>
                  <EmptyMedia variant="icon">
                    <Icon name="i-lucide-folder-code" class="size-6" />
                  </EmptyMedia>
                  <EmptyTitle>No Notifications</EmptyTitle>
                  <EmptyDescription>You have no new alerts.</EmptyDescription>
                </EmptyHeader>
              </Empty>
            </template>
          </DropdownMenuContent>
        </DropdownMenu>
      </div>
    </div>
    <div class="ml-auto">
      <slot />
    </div>
  </header>
</template>

<style scoped>

</style>
