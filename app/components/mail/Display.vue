<script lang="ts" setup>
import type { Mail } from './data/mails'
import { format } from 'date-fns'
import { Archive, ArchiveX, Forward, MailOpen, MoreHorizontal, Reply, ReplyAll, Trash2 } from 'lucide-vue-next'
import { computed } from 'vue'

interface MailDisplayProps {
  mail: Mail | undefined
}

const props = defineProps<MailDisplayProps>()

const emit = defineEmits(['close'])

const mailFallbackName = computed(() => {
  return props.mail?.name
    .split(' ')
    .map(chunk => chunk[0])
    .join('')
})

const avatarColors = ['#0078D4', '#8764B8', '#107C10', '#CA5010', '#D13438', '#5C2D91']
const avatarColor = computed(() => {
  if (!props.mail)
    return avatarColors[0]
  const sum = props.mail.name.split('').reduce((acc, c) => acc + c.charCodeAt(0), 0)
  return avatarColors[sum % avatarColors.length]
})
</script>

<template>
  <div class="flex h-full flex-col">
    <div class="flex h-[52px] shrink-0 items-center gap-1 border-b px-3">
      <Tooltip>
        <TooltipTrigger as-child>
          <Button variant="ghost" size="icon" class="size-8" :disabled="!mail">
            <Archive class="size-4" />
            <span class="sr-only">Archive</span>
          </Button>
        </TooltipTrigger>
        <TooltipContent>Archive</TooltipContent>
      </Tooltip>
      <Tooltip>
        <TooltipTrigger as-child>
          <Button variant="ghost" size="icon" class="size-8" :disabled="!mail">
            <ArchiveX class="size-4" />
            <span class="sr-only">Move to junk</span>
          </Button>
        </TooltipTrigger>
        <TooltipContent>Move to junk</TooltipContent>
      </Tooltip>
      <Tooltip>
        <TooltipTrigger as-child>
          <Button variant="ghost" size="icon" class="size-8" :disabled="!mail" @click="emit('close')">
            <Trash2 class="size-4" />
            <span class="sr-only">Delete</span>
          </Button>
        </TooltipTrigger>
        <TooltipContent>Delete</TooltipContent>
      </Tooltip>

      <div class="ml-auto flex items-center gap-1.5">
        <Button variant="outline" size="sm" class="h-8 gap-1.5 text-xs" :disabled="!mail">
          <Reply class="size-3.5" /> Reply
        </Button>
        <Button variant="outline" size="sm" class="h-8 gap-1.5 text-xs" :disabled="!mail">
          <ReplyAll class="size-3.5" /> Reply all
        </Button>
        <Button variant="outline" size="sm" class="h-8 gap-1.5 text-xs" :disabled="!mail">
          <Forward class="size-3.5" /> Forward
        </Button>
        <DropdownMenu>
          <DropdownMenuTrigger as-child>
            <Button variant="ghost" size="icon" class="size-8" :disabled="!mail">
              <MoreHorizontal class="size-4" />
              <span class="sr-only">More</span>
            </Button>
          </DropdownMenuTrigger>
          <DropdownMenuContent align="end">
            <DropdownMenuItem>Mark as unread</DropdownMenuItem>
            <DropdownMenuItem>Flag</DropdownMenuItem>
            <DropdownMenuItem>Categorize</DropdownMenuItem>
            <DropdownMenuItem>Mute thread</DropdownMenuItem>
          </DropdownMenuContent>
        </DropdownMenu>
      </div>
    </div>

    <div v-if="mail" class="flex flex-1 flex-col overflow-hidden">
      <ScrollArea class="flex-1">
        <div class="px-6 py-4">
          <h1 class="text-xl font-semibold">
            {{ mail.subject }}
          </h1>
          <div class="mt-3 flex items-start gap-3">
            <Avatar class="size-9">
              <AvatarFallback :style="{ backgroundColor: avatarColor }" class="text-xs font-medium text-white">
                {{ mailFallbackName }}
              </AvatarFallback>
            </Avatar>
            <div class="min-w-0 flex-1">
              <div class="flex items-baseline gap-2">
                <span class="font-semibold">{{ mail.name }}</span>
                <span v-if="mail.date" class="ml-auto shrink-0 text-xs text-muted-foreground">
                  {{ format(new Date(mail.date), "EEE M/d/yyyy h:mm a") }}
                </span>
              </div>
              <div class="text-xs text-muted-foreground">
                To: me &lt;{{ mail.email }}&gt;
              </div>
            </div>
          </div>
          <Separator class="my-4" />
          <div class="whitespace-pre-wrap text-sm leading-relaxed">
            {{ mail.text }}
          </div>
        </div>
      </ScrollArea>

      <div class="shrink-0 border-t p-3">
        <div class="rounded-lg border">
          <Textarea
            class="min-h-[70px] resize-none border-0 focus-visible:ring-0"
            :placeholder="`Reply to ${mail.name}...`"
          />
          <div class="flex items-center gap-2 border-t px-3 py-2">
            <Button size="sm" class="gap-1.5 bg-[#0078D4] text-white hover:bg-[#106EBE]">
              <Reply class="size-3.5" /> Send
            </Button>
            <Label class="ml-auto flex items-center gap-2 text-xs font-normal text-muted-foreground">
              <Switch id="mute" aria-label="Mute thread" /> Mute this thread
            </Label>
          </div>
        </div>
      </div>
    </div>
    <div v-else class="flex flex-1 flex-col items-center justify-center gap-2 text-muted-foreground">
      <MailOpen class="size-10 opacity-40" />
      <p class="text-sm">
        No message selected
      </p>
    </div>
  </div>
</template>
