<script setup lang="ts">
import type { ColumnDef } from '@tanstack/vue-table'

interface Account { id: string; name: string; owner_user_id: string | null; created_at: string }

const api = useApi()
const accounts = ref<Account[]>([])
const newAccountName = ref('')
const creating = ref(false)

async function load() {
  accounts.value = await api<Account[]>('/sales/accounts') ?? []
}

async function createAccount() {
  if (!newAccountName.value.trim()) return
  creating.value = true
  try {
    await api('/sales/accounts', { method: 'POST', body: { name: newAccountName.value } })
    newAccountName.value = ''
    await load()
  } finally {
    creating.value = false
  }
}

const columns: ColumnDef<Account, any>[] = [
  { accessorKey: 'name', header: 'Name', enableSorting: true },
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
    <div class="flex items-center justify-between">
      <div>
        <h1 class="text-2xl font-semibold tracking-tight">Accounts</h1>
        <p class="text-muted-foreground">Companies and organizations you do business with.</p>
      </div>
    </div>

    <Card>
      <CardHeader>
        <CardTitle class="text-base">New account</CardTitle>
      </CardHeader>
      <CardContent>
        <form class="flex gap-2" @submit.prevent="createAccount">
          <Input v-model="newAccountName" placeholder="Account name" class="max-w-sm" />
          <Button type="submit" :disabled="creating">Add</Button>
        </form>
      </CardContent>
    </Card>

    <BaseDataTable :columns="columns" :data="accounts" search-placeholder="Filter accounts…" />
  </div>
</template>
