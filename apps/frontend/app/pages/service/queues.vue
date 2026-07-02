<script setup lang="ts">
import type { ColumnDef } from '@tanstack/vue-table'

interface Queue { id: string; name: string }

const api = useApi()
const queues = ref<Queue[]>([])
const newName = ref('')
const creating = ref(false)

async function load() {
  queues.value = await api<Queue[]>('/service/queues') ?? []
}

async function createQueue() {
  if (!newName.value.trim()) return
  creating.value = true
  try {
    await api('/service/queues', { method: 'POST', body: { name: newName.value } })
    newName.value = ''
    await load()
  } finally {
    creating.value = false
  }
}

const columns: ColumnDef<Queue, any>[] = [
  { accessorKey: 'name', header: 'Name', enableSorting: true },
]

onMounted(load)
</script>

<template>
  <div class="flex flex-col gap-6">
    <div>
      <h1 class="text-2xl font-semibold tracking-tight">Queues</h1>
      <p class="text-muted-foreground">Case routing destinations, typically one per client engagement.</p>
    </div>

    <Card>
      <CardHeader>
        <CardTitle class="text-base">New queue</CardTitle>
      </CardHeader>
      <CardContent>
        <form class="flex gap-2" @submit.prevent="createQueue">
          <Input v-model="newName" placeholder="Queue name" class="max-w-sm" />
          <Button type="submit" :disabled="creating">Add</Button>
        </form>
      </CardContent>
    </Card>

    <BaseDataTable :columns="columns" :data="queues" search-placeholder="Filter queues…" />
  </div>
</template>
