<script setup lang="ts">
import type { ColumnDef } from '@tanstack/vue-table'

interface Contact { id: string; first_name: string | null; last_name: string | null; email: string | null; phone: string | null }

const api = useApi()
const contacts = ref<Contact[]>([])
const newFirst = ref('')
const newLast = ref('')
const newEmail = ref('')
const newPhone = ref('')
const creating = ref(false)

async function load() {
  contacts.value = await api<Contact[]>('/sales/contacts') ?? []
}

async function createContact() {
  if (!newFirst.value.trim() && !newLast.value.trim()) return
  creating.value = true
  try {
    await api('/sales/contacts', {
      method: 'POST',
      body: { first_name: newFirst.value, last_name: newLast.value, email: newEmail.value, phone: newPhone.value },
    })
    newFirst.value = ''
    newLast.value = ''
    newEmail.value = ''
    newPhone.value = ''
    await load()
  } finally {
    creating.value = false
  }
}

const columns: ColumnDef<Contact, any>[] = [
  {
    id: 'name',
    header: 'Name',
    enableSorting: true,
    accessorFn: (row) => [row.first_name, row.last_name].filter(Boolean).join(' '),
  },
  { accessorKey: 'email', header: 'Email' },
  { accessorKey: 'phone', header: 'Phone' },
]

onMounted(load)
</script>

<template>
  <div class="flex flex-col gap-6">
    <div>
      <h1 class="text-2xl font-semibold tracking-tight">Contacts</h1>
      <p class="text-muted-foreground">People you deal with, optionally tied to an account.</p>
    </div>

    <Card>
      <CardHeader>
        <CardTitle class="text-base">New contact</CardTitle>
      </CardHeader>
      <CardContent>
        <form class="flex flex-wrap gap-2" @submit.prevent="createContact">
          <Input v-model="newFirst" placeholder="First name" class="max-w-40" />
          <Input v-model="newLast" placeholder="Last name" class="max-w-40" />
          <Input v-model="newEmail" placeholder="Email" type="email" class="max-w-56" />
          <Input v-model="newPhone" placeholder="Phone" class="max-w-40" />
          <Button type="submit" :disabled="creating">Add</Button>
        </form>
      </CardContent>
    </Card>

    <BaseDataTable :columns="columns" :data="contacts" search-placeholder="Filter contacts…" />
  </div>
</template>
