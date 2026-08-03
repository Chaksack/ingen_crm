<script setup lang="ts">
import { toast } from 'vue-sonner'

const requestUrl = useRequestURL()
const supportLink = computed(() => `${requestUrl.origin}/support`)

const table = ref<{ refresh: () => Promise<void> } | null>(null)
const selectedTicketId = ref<string | null>(null)
const isSheetOpen = ref(false)

interface TicketRow {
  id: string
}

function openTicket(ticket: TicketRow) {
  selectedTicketId.value = ticket.id
  isSheetOpen.value = true
}

function onTicketChanged() {
  table.value?.refresh()
}

async function copySupportLink() {
  try {
    await navigator.clipboard.writeText(supportLink.value)
    toast.success('Support link copied to clipboard')
  }
  catch {
    toast.error('Could not copy link')
  }
}
</script>

<template>
  <div class="w-full flex flex-col items-stretch gap-4">
    <div class="flex flex-wrap items-end justify-between gap-2">
      <div>
        <h2 class="text-2xl font-bold tracking-tight">
          Customer Support
        </h2>
        <p class="text-muted-foreground">
          Tickets submitted by clients through your public support link.
        </p>
      </div>
      <Button variant="outline" @click="copySupportLink">
        <Icon name="i-lucide-link" />
        Copy Support Link
      </Button>
    </div>

    <CustomerSupportTicketTable ref="table" @select="openTicket" />

    <CustomerSupportTicketDetailSheet
      v-model:open="isSheetOpen"
      :ticket-id="selectedTicketId"
      @changed="onTicketChanged"
    />
  </div>
</template>
