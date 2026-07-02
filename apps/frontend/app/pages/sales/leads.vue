<script setup lang="ts">
import { h } from 'vue'
import type { ColumnDef } from '@tanstack/vue-table'

interface Lead { id: string; topic: string; status: string; created_at: string }

const api = useApi()
const leads = ref<Lead[]>([])
const newTopic = ref('')
const creating = ref(false)

async function load() {
  leads.value = await api<Lead[]>('/sales/leads') ?? []
}

async function createLead() {
  if (!newTopic.value.trim()) return
  creating.value = true
  try {
    await api('/sales/leads', { method: 'POST', body: { topic: newTopic.value } })
    newTopic.value = ''
    await load()
  } finally {
    creating.value = false
  }
}

const columns: ColumnDef<Lead, any>[] = [
  { accessorKey: 'topic', header: 'Topic', enableSorting: true },
  {
    accessorKey: 'status',
    header: 'Status',
    enableSorting: true,
    cell: ({ row }) => h(resolveComponent('Badge'), { variant: 'outline' }, () => row.original.status),
  },
  {
    accessorKey: 'created_at',
    header: 'Created',
    enableSorting: true,
    cell: ({ row }) => new Date(row.original.created_at).toLocaleString(),
  },
]

onMounted(load)
</script>

<template>
  <div class="flex flex-col gap-6">
    <div>
      <h1 class="text-2xl font-semibold tracking-tight">Leads</h1>
      <p class="text-muted-foreground">Inbound interest, not yet qualified into an opportunity.</p>
    </div>

    <Card>
      <CardHeader>
        <CardTitle class="text-base">New lead</CardTitle>
      </CardHeader>
      <CardContent>
        <form class="flex gap-2" @submit.prevent="createLead">
          <Input v-model="newTopic" placeholder="Lead topic" class="max-w-sm" />
          <Button type="submit" :disabled="creating">Add</Button>
        </form>
      </CardContent>
    </Card>

    <BaseDataTable :columns="columns" :data="leads" search-placeholder="Filter leads…" />
  </div>
</template>
