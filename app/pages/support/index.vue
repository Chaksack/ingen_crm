<script setup lang="ts">
import { Loader2, Mail, MessageCircle } from 'lucide-vue-next'
import { toast } from 'vue-sonner'

definePageMeta({
  layout: 'blank',
})

interface ChatSession {
  id: string
  accessToken: string
  ticketNumber: string
  description: string
}

const STORAGE_KEY = 'ingenicx-support-chat'

const isLoading = ref(false)
const submittedTicket = ref<string | null>(null)
const chatSession = ref<ChatSession | null>(null)

const form = reactive({
  name: '',
  email: '',
  phone: '',
  company: '',
  subject: '',
  description: '',
  category: '',
  priority: 'medium' as 'low' | 'medium' | 'high' | 'urgent',
  preferredContact: 'email' as 'chat' | 'email',
})

onMounted(() => {
  const saved = localStorage.getItem(STORAGE_KEY)
  if (saved) {
    try {
      chatSession.value = JSON.parse(saved)
    }
    catch {
      localStorage.removeItem(STORAGE_KEY)
    }
  }
})

function startNewConversation() {
  localStorage.removeItem(STORAGE_KEY)
  chatSession.value = null
  submittedTicket.value = null
  Object.assign(form, { name: '', email: '', phone: '', company: '', subject: '', description: '', category: '', priority: 'medium', preferredContact: 'email' })
}

async function onSubmit(event: Event) {
  event.preventDefault()
  if (!form.name || !form.email || !form.subject || !form.description) {
    toast.error('Please fill in your name, email, subject, and description')
    return
  }

  isLoading.value = true
  try {
    const result = await $fetch<{ ticketNumber: string, id: string, accessToken: string, preferredContact: 'chat' | 'email' }>('/api/support/tickets', {
      method: 'POST',
      body: form,
    })

    if (result.preferredContact === 'chat') {
      const session: ChatSession = {
        id: result.id,
        accessToken: result.accessToken,
        ticketNumber: result.ticketNumber,
        description: form.description,
      }
      chatSession.value = session
      localStorage.setItem(STORAGE_KEY, JSON.stringify(session))
    }
    else {
      submittedTicket.value = result.ticketNumber
    }
  }
  catch (err: any) {
    toast.error(err?.data?.statusMessage || 'Failed to submit ticket. Please try again.')
  }
  finally {
    isLoading.value = false
  }
}
</script>

<template>
  <div class="flex flex-col items-center justify-center gap-6 bg-muted p-6 min-h-svh md:p-10">
    <div class="w-full flex flex-col gap-6" :class="chatSession ? 'max-w-xl' : 'max-w-lg'">
      <div class="flex items-center self-center gap-2 font-medium">
        <h1 class="font-bold text-2xl">
          Ingenicx
        </h1>
      </div>

      <template v-if="chatSession">
        <CustomerSupportPublicChat
          :ticket-id="chatSession.id"
          :access-token="chatSession.accessToken"
          :description="chatSession.description"
          @end-chat="startNewConversation"
        />
      </template>

      <Card v-else-if="submittedTicket">
        <CardHeader class="text-center">
          <div class="mx-auto mb-2 flex size-12 items-center justify-center rounded-full bg-primary/10">
            <Icon name="i-lucide-check" class="size-6 text-primary" />
          </div>
          <CardTitle class="text-xl">
            Ticket submitted
          </CardTitle>
          <CardDescription>
            Your ticket <span class="font-medium text-foreground">{{ submittedTicket }}</span> has been received.
            Our support team will get back to you by email shortly.
          </CardDescription>
        </CardHeader>
      </Card>

      <Card v-else>
        <CardHeader class="text-center">
          <CardTitle class="text-xl">
            Contact Support
          </CardTitle>
          <CardDescription>
            Tell us what's going on and we'll get back to you as soon as possible.
          </CardDescription>
        </CardHeader>
        <CardContent>
          <form class="grid gap-4" @submit="onSubmit">
            <div class="grid gap-2">
              <Label>How would you like to reach us?</Label>
              <div class="grid grid-cols-2 gap-3">
                <button
                  type="button"
                  class="flex flex-col items-center gap-2 rounded-lg border-2 p-4 text-sm transition-colors"
                  :class="form.preferredContact === 'chat' ? 'border-primary bg-primary/5' : 'border-muted hover:border-muted-foreground/30'"
                  @click="form.preferredContact = 'chat'"
                >
                  <MessageCircle class="h-5 w-5" />
                  <span class="font-medium">Live Chat</span>
                  <span class="text-xs text-muted-foreground text-center">Talk to an agent right now</span>
                </button>
                <button
                  type="button"
                  class="flex flex-col items-center gap-2 rounded-lg border-2 p-4 text-sm transition-colors"
                  :class="form.preferredContact === 'email' ? 'border-primary bg-primary/5' : 'border-muted hover:border-muted-foreground/30'"
                  @click="form.preferredContact = 'email'"
                >
                  <Mail class="h-5 w-5" />
                  <span class="font-medium">Email</span>
                  <span class="text-xs text-muted-foreground text-center">We'll follow up by email</span>
                </button>
              </div>
            </div>

            <div class="grid grid-cols-2 gap-4">
              <div class="grid gap-2">
                <Label for="name">Your Name</Label>
                <Input id="name" v-model="form.name" :disabled="isLoading" />
              </div>
              <div class="grid gap-2">
                <Label for="email">Email</Label>
                <Input id="email" v-model="form.email" type="email" :disabled="isLoading" />
              </div>
            </div>
            <div class="grid grid-cols-2 gap-4">
              <div class="grid gap-2">
                <Label for="phone">Phone (optional)</Label>
                <Input id="phone" v-model="form.phone" :disabled="isLoading" />
              </div>
              <div class="grid gap-2">
                <Label for="company">Company (optional)</Label>
                <Input id="company" v-model="form.company" :disabled="isLoading" />
              </div>
            </div>
            <div class="grid gap-2">
              <Label for="subject">Subject</Label>
              <Input id="subject" v-model="form.subject" placeholder="Brief summary of the issue" :disabled="isLoading" />
            </div>
            <div class="grid grid-cols-2 gap-4">
              <div class="grid gap-2">
                <Label for="category">Category (optional)</Label>
                <Input id="category" v-model="form.category" placeholder="e.g. Billing, Technical" :disabled="isLoading" />
              </div>
              <div class="grid gap-2">
                <Label for="priority">Priority</Label>
                <Select v-model="form.priority">
                  <SelectTrigger id="priority" class="w-full">
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="low">
                      Low
                    </SelectItem>
                    <SelectItem value="medium">
                      Medium
                    </SelectItem>
                    <SelectItem value="high">
                      High
                    </SelectItem>
                    <SelectItem value="urgent">
                      Urgent
                    </SelectItem>
                  </SelectContent>
                </Select>
              </div>
            </div>
            <div class="grid gap-2">
              <Label for="description">Description</Label>
              <Textarea id="description" v-model="form.description" rows="5" placeholder="Please describe the issue in detail" :disabled="isLoading" />
            </div>
            <Button type="submit" class="w-full" :disabled="isLoading">
              <Loader2 v-if="isLoading" class="mr-2 h-4 w-4 animate-spin" />
              {{ form.preferredContact === 'chat' ? 'Start Chat' : 'Submit Ticket' }}
            </Button>
          </form>
        </CardContent>
      </Card>

      <div class="text-center text-balance text-xs text-muted-foreground">
        © 2026 Ingenicx. All rights reserved.
      </div>
    </div>
  </div>
</template>
