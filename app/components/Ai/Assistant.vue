<script setup lang="ts">
import { ref, computed, nextTick } from 'vue'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { ScrollArea } from '@/components/ui/scroll-area'
import { Avatar, AvatarFallback, AvatarImage } from '@/components/ui/avatar'
import { Card, CardContent } from '@/components/ui/card'
import { Skeleton } from '@/components/ui/skeleton'
import { Send, Bot, User, Sparkles } from 'lucide-vue-next'

interface Message {
  id: number
  role: 'user' | 'assistant'
  content: string
  timestamp: Date
}

const messages = ref<Message[]>([
  {
    id: 1,
    role: 'assistant',
    content: 'Hello! I\'m your AI assistant. How can I help you today?',
    timestamp: new Date()
  }
])

const userInput = ref('')
const isLoading = ref(false)
const scrollAreaRef = ref<HTMLElement | null>(null)

// Suggested prompts for quick actions
const suggestedPrompts = [
  'Analyze customer credit trends',
  'Generate a financial report',
  'Check pending loan applications',
  'Summarize recent transactions'
]

// Scroll to bottom of messages
const scrollToBottom = async () => {
  await nextTick()
  if (scrollAreaRef.value) {
    const scrollContainer = scrollAreaRef.value.querySelector('[data-radix-scroll-area-viewport]')
    if (scrollContainer) {
      scrollContainer.scrollTop = scrollContainer.scrollHeight
    }
  }
}

// Handle sending a message
const sendMessage = async () => {
  if (!userInput.value.trim() || isLoading.value) return

  const userMessage: Message = {
    id: Date.now(),
    role: 'user',
    content: userInput.value,
    timestamp: new Date()
  }

  messages.value.push(userMessage)
  const query = userInput.value
  userInput.value = ''
  
  scrollToBottom()

  // Show loading state
  isLoading.value = true

  try {
    // Simulate AI response (replace with actual API call)
    await new Promise(resolve => setTimeout(resolve, 1500))

    const assistantMessage: Message = {
      id: Date.now() + 1,
      role: 'assistant',
      content: `I understand you're asking about: "${query}". Here's what I can help you with:\n\n1. I can analyze customer data and credit scores\n2. Generate detailed financial reports\n3. Provide insights on loan applications\n4. Help with data visualization\n\nWhat specific aspect would you like to explore?`,
      timestamp: new Date()
    }

    messages.value.push(assistantMessage)
    scrollToBottom()
  } catch (error) {
    console.error('Error sending message:', error)
  } finally {
    isLoading.value = false
  }
}

// Handle suggested prompt click
const useSuggestedPrompt = (prompt: string) => {
  userInput.value = prompt
  sendMessage()
}

// Format timestamp
const formatTime = (date: Date) => {
  return date.toLocaleTimeString('en-US', { 
    hour: '2-digit', 
    minute: '2-digit' 
  })
}

// Handle Enter key
const handleKeyPress = (event: KeyboardEvent) => {
  if (event.key === 'Enter' && !event.shiftKey) {
    event.preventDefault()
    sendMessage()
  }
}
</script>

<template>
  <div class="flex flex-col h-[calc(100vh-8rem)] gap-4">
    <!-- Messages Area -->
    <ScrollArea ref="scrollAreaRef" class="flex-1 pr-4">
      <div class="space-y-4">
        <div
          v-for="message in messages"
          :key="message.id"
          :class="[
            'flex gap-3',
            message.role === 'user' ? 'justify-end' : 'justify-start'
          ]"
        >
          <!-- Assistant Avatar -->
          <Avatar v-if="message.role === 'assistant'" class="h-8 w-8 mt-1">
            <AvatarFallback class="bg-primary text-primary-foreground">
              <Bot class="h-4 w-4" />
            </AvatarFallback>
          </Avatar>

          <!-- Message Content -->
          <div
            :class="[
              'max-w-[80%] rounded-lg px-4 py-2',
              message.role === 'user'
                ? 'bg-primary text-primary-foreground'
                : 'bg-muted'
            ]"
          >
            <p class="text-sm whitespace-pre-wrap">{{ message.content }}</p>
            <span class="text-[10px] opacity-70 mt-1 block">
              {{ formatTime(message.timestamp) }}
            </span>
          </div>

          <!-- User Avatar -->
          <Avatar v-if="message.role === 'user'" class="h-8 w-8 mt-1">
            <AvatarFallback class="bg-secondary">
              <User class="h-4 w-4" />
            </AvatarFallback>
          </Avatar>
        </div>

        <!-- Loading Indicator -->
        <div v-if="isLoading" class="flex gap-3 justify-start">
          <Avatar class="h-8 w-8 mt-1">
            <AvatarFallback class="bg-primary text-primary-foreground">
              <Bot class="h-4 w-4" />
            </AvatarFallback>
          </Avatar>
          <div class="max-w-[80%] rounded-lg px-4 py-3 bg-muted">
            <div class="flex gap-1">
              <div class="w-2 h-2 rounded-full bg-primary/60 animate-bounce [animation-delay:-0.3s]"></div>
              <div class="w-2 h-2 rounded-full bg-primary/60 animate-bounce [animation-delay:-0.15s]"></div>
              <div class="w-2 h-2 rounded-full bg-primary/60 animate-bounce"></div>
            </div>
          </div>
        </div>
      </div>
    </ScrollArea>

    <!-- Suggested Prompts (shown when no messages except initial) -->
    <div v-if="messages.length === 1" class="space-y-2">
      <p class="text-xs text-muted-foreground flex items-center gap-1">
        <Sparkles class="h-3 w-3" />
        Suggested prompts
      </p>
      <div class="grid grid-cols-1 gap-2">
        <Button
          v-for="prompt in suggestedPrompts"
          :key="prompt"
          variant="outline"
          size="sm"
          class="justify-start text-left h-auto py-2 px-3"
          @click="useSuggestedPrompt(prompt)"
        >
          <span class="text-xs">{{ prompt }}</span>
        </Button>
      </div>
    </div>

    <!-- Input Area -->
    <div class="flex gap-2 pt-2 border-t">
      <Input
        v-model="userInput"
        placeholder="Ask me anything..."
        class="flex-1"
        :disabled="isLoading"
        @keypress="handleKeyPress"
      />
      <Button
        size="icon"
        :disabled="!userInput.trim() || isLoading"
        @click="sendMessage"
      >
        <Send class="h-4 w-4" />
      </Button>
    </div>

    <!-- Info Footer -->
    <div class="text-[10px] text-muted-foreground text-center">
      AI can make mistakes. Please verify important information.
    </div>
  </div>
</template>

<style scoped>
@keyframes bounce {
  0%, 100% {
    transform: translateY(0);
  }
  50% {
    transform: translateY(-4px);
  }
}

.animate-bounce {
  animation: bounce 1s ease-in-out infinite;
}
</style>