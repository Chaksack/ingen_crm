<script lang="ts" setup>
import type { Mail } from './data/mails'
import { formatDistanceToNow } from 'date-fns'
import { cn } from '@/lib/utils'

interface MailListProps {
  items: Mail[]
}

defineProps<MailListProps>()
const selectedMail = defineModel<string>('selectedMail', { required: false })

const labelColors: Record<string, string> = {
  work: 'bg-[#0078D4]',
  important: 'bg-[#D13438]',
  meeting: 'bg-[#8764B8]',
  personal: 'bg-[#107C10]',
  budget: 'bg-[#CA5010]',
}

function getLabelColor(label: string) {
  return labelColors[label.toLowerCase()] ?? 'bg-muted-foreground'
}
</script>

<template>
  <ScrollArea class="flex-1">
    <div class="flex flex-col">
      <TransitionGroup name="list" appear>
        <button
          v-for="item of items"
          :key="item.id"
          :class="cn(
            'relative flex w-full flex-col gap-1 border-b px-4 py-2.5 text-left transition-colors hover:bg-accent/60',
            selectedMail === item.id && 'bg-[#EFF6FC] hover:bg-[#EFF6FC] dark:bg-[#0078D4]/15 dark:hover:bg-[#0078D4]/15',
          )"
          @click="selectedMail = item.id"
        >
          <span
            v-if="!item.read || selectedMail === item.id"
            class="absolute inset-y-0 left-0 w-[3px] bg-[#0078D4]"
          />
          <div class="flex items-baseline gap-2">
            <span :class="cn('truncate text-sm', !item.read ? 'font-semibold text-foreground' : 'font-medium text-foreground/90')">
              {{ item.name }}
            </span>
            <span class="ml-auto shrink-0 text-[11px] text-muted-foreground">
              {{ formatDistanceToNow(new Date(item.date), { addSuffix: true }) }}
            </span>
          </div>
          <div class="flex items-baseline gap-1 text-xs">
            <span :class="cn('shrink-0 truncate', !item.read ? 'font-medium text-foreground' : 'text-foreground/80')">
              {{ item.subject }}
            </span>
            <span class="truncate text-muted-foreground">
              — {{ item.text.substring(0, 80) }}
            </span>
          </div>
          <div v-if="item.labels.length" class="flex items-center gap-2.5 pt-0.5">
            <span
              v-for="label of item.labels"
              :key="label"
              class="flex items-center gap-1 text-[10px] text-muted-foreground"
            >
              <span :class="cn('size-1.5 rounded-full', getLabelColor(label))" />
              {{ label }}
            </span>
          </div>
        </button>
      </TransitionGroup>
      <div v-if="items.length === 0" class="p-8 text-center text-sm text-muted-foreground">
        No messages
      </div>
    </div>
  </ScrollArea>
</template>

<style scoped>
.list-move,
.list-enter-active,
.list-leave-active {
  transition: all 0.3s ease;
}

.list-enter-from,
.list-leave-to {
  opacity: 0;
}

.list-leave-active {
  position: absolute;
}
</style>
