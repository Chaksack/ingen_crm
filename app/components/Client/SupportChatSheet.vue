<script setup lang="ts">
import { ref } from 'vue'
import { Sheet, SheetContent, SheetHeader, SheetTitle, SheetDescription, SheetFooter } from '@/components/ui/sheet'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import type { Task } from '@/components/tasks/data/schema'
import { statuses } from '@/components/tasks/data/data'
import { Select, SelectContent, SelectGroup, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'

interface SupportChatSheetProps {
  task: Task | null
  open: boolean
}

const props = defineProps<SupportChatSheetProps>()
const emit = defineEmits(['update:open', 'update:task'])

const newMessage = ref('')

function sendMessage() {
  // Logic to send a message
  console.log('Sending message:', newMessage.value)
  newMessage.value = ''
}

function onOpenUpdate(value: boolean) {
  emit('update:open', value)
}

function updateTaskStatus(status: string) {
  if (props.task) {
    const updatedTask = { ...props.task, status };
    emit('update:task', updatedTask);
  }
}
</script>

<template>
  <Sheet :open="open" @update:open="onOpenUpdate">
    <SheetContent class="w-full max-w-md px-2 py-4 sm:max-w-lg rounded-l-lg overflow-y-auto">
      <SheetHeader>
        <SheetTitle>Chat with Customer</SheetTitle>
        <SheetDescription v-if="task">
          Support case for task: {{ task.id }}
        </SheetDescription>
      </SheetHeader>
      <div class="py-4 h-[calc(100%-200px)] overflow-y-auto">
        <!-- Chat history will go here -->
        <div class="flex flex-col gap-4">
          <div class="flex gap-2">
            <div class="w-10 h-10 rounded-full bg-muted flex items-center justify-center font-bold">C</div>
            <div class="p-2 rounded-lg bg-muted">
              <p class="text-sm">
                Hello, I'm having an issue with my account.
              </p>
            </div>
          </div>
          <div class="flex gap-2 flex-row-reverse">
            <div class="w-10 h-10 rounded-full bg-primary text-primary-foreground flex items-center justify-center font-bold">A</div>
            <div class="p-2 rounded-lg bg-primary text-primary-foreground">
              <p class="text-sm">
                Hi there! I can help with that. What seems to be the problem?
              </p>
            </div>
          </div>
        </div>
      </div>
      <SheetFooter>
        <div class="flex flex-col w-full gap-4">
           <div class="flex items-center gap-2">
            <Input v-model="newMessage" placeholder="Type your message..." @keyup.enter="sendMessage" />
            <Button @click="sendMessage">
              Send
            </Button>
          </div>
          <div class="flex items-center gap-2">
            <Label for="status" class="shrink-0">Status</Label>
            <Select v-if="task" :model-value="task.status" @update:model-value="updateTaskStatus">
              <SelectTrigger>
                <SelectValue placeholder="Select status" />
              </SelectTrigger>
              <SelectContent>
                <SelectGroup>
                  <SelectItem v-for="statusItem in statuses" :key="statusItem.value" :value="statusItem.value">
                    {{ statusItem.label }}
                  </SelectItem>
                </SelectGroup>
              </SelectContent>
            </Select>
            <Button variant="outline" @click="onOpenUpdate(false)">
              Close
            </Button>
          </div>
        </div>
      </SheetFooter>
    </SheetContent>
  </Sheet>
</template>
