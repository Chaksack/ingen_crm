<script setup lang="ts">
import { Search } from 'lucide-vue-next'
import { computed, ref } from 'vue'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { Skeleton } from '@/components/ui/skeleton'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table'

interface ClientItem {
  id: string
  name: string
  contactPerson?: string
  email: string
  phone?: string
  status: 'active' | 'inactive'
  address?: string
  notes?: string
}

const { data: clientsData, pending: isLoading, refresh } = useFetch<ClientItem[]>('/api/clients')
const clients = computed(() => clientsData.value ?? [])

const searchQuery = ref('')
const statusFilter = ref<'all' | 'active' | 'inactive'>('all')

const filtered = computed(() => {
  let items = clients.value
  const q = searchQuery.value.trim().toLowerCase()
  if (q) {
    items = items.filter(i =>
      [i.name, i.contactPerson, i.email, i.phone, i.status, i.address, i.notes]
        .filter(Boolean)
        .some(v => String(v).toLowerCase().includes(q)),
    )
  }
  if (statusFilter.value !== 'all')
    items = items.filter(i => i.status === statusFilter.value)
  return items
})

function clearFilters() {
  searchQuery.value = ''
  statusFilter.value = 'all'
}

defineExpose({ refresh })
</script>

<template>
  <div class="space-y-4">
    <div class="flex flex-col gap-4">
      <div class="flex flex-col md:flex-row gap-4 items-start md:items-center justify-between">
        <div class="relative flex-1 max-w-sm">
          <Search class="absolute left-2 top-2.5 h-4 w-4 text-muted-foreground" />
          <Input
            id="search"
            v-model="searchQuery"
            placeholder="Search clients..."
            class="pl-8"
          />
        </div>

        <div class="flex gap-2 flex-wrap">
          <Select v-model="statusFilter">
            <SelectTrigger id="status" class="w-[180px]">
              <SelectValue placeholder="Filter by status" />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value="all">
                All Status
              </SelectItem>
              <SelectItem value="active">
                Active
              </SelectItem>
              <SelectItem value="inactive">
                Inactive
              </SelectItem>
            </SelectContent>
          </Select>

          <Button v-if="searchQuery || statusFilter !== 'all'" variant="outline" @click="clearFilters">
            Clear Filters
          </Button>
        </div>
      </div>

      <div class="text-sm text-muted-foreground">
        Showing {{ filtered.length }} of {{ clients.length }} results
        <span v-if="searchQuery || statusFilter !== 'all'">
          (filtered)
        </span>
      </div>
    </div>

    <div class="rounded-md border">
      <Table>
        <TableHeader>
          <TableRow>
            <TableHead>Client</TableHead>
            <TableHead>Contact</TableHead>
            <TableHead>Email</TableHead>
            <TableHead>Phone</TableHead>
            <TableHead>Status</TableHead>
            <TableHead class="text-right">
              Actions
            </TableHead>
          </TableRow>
        </TableHeader>
        <TableBody v-if="isLoading">
          <TableRow v-for="i in 5" :key="`skeleton-${i}`">
            <TableCell>
              <Skeleton class="h-4 w-32" />
            </TableCell>
            <TableCell>
              <Skeleton class="h-4 w-24" />
            </TableCell>
            <TableCell>
              <Skeleton class="h-4 w-40" />
            </TableCell>
            <TableCell>
              <Skeleton class="h-4 w-28" />
            </TableCell>
            <TableCell>
              <Skeleton class="h-6 w-16 rounded-full" />
            </TableCell>
            <TableCell class="text-right">
              <Skeleton class="h-8 w-16 ml-auto rounded-lg" />
            </TableCell>
          </TableRow>
        </TableBody>
        <TableBody v-else>
          <TableRow v-for="item in filtered" :key="item.id">
            <TableCell class="font-medium">
              {{ item.name }}
            </TableCell>
            <TableCell>
              {{ item.contactPerson }}
            </TableCell>
            <TableCell>
              {{ item.email }}
            </TableCell>
            <TableCell>
              {{ item.phone }}
            </TableCell>
            <TableCell>
              <Badge :variant="item.status === 'active' ? 'default' : 'secondary'">
                {{ item.status }}
              </Badge>
            </TableCell>
            <TableCell class="text-right space-x-2">
              <NuxtLink :to="`/clients/view/${item.id}`">
                <Button size="sm" variant="outline">
                  View
                </Button>
              </NuxtLink>
            </TableCell>
          </TableRow>
          <TableRow v-if="filtered.length === 0">
            <TableCell colspan="6" class="text-center text-muted-foreground">
              No clients found.
            </TableCell>
          </TableRow>
        </TableBody>
      </Table>
    </div>
  </div>
</template>
