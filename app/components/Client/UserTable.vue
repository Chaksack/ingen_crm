<script setup lang="ts">
import { Search } from 'lucide-vue-next'
import { computed, ref } from 'vue'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table'

interface ClientItem {
  id: number
  name: string
  contactPerson: string
  email: string
  phone: string
  status: 'active' | 'inactive'
  address?: string
  notes?: string
}

const clients = ref<ClientItem[]>([
  {
    id: 1,
    name: 'Wilmar Africa',
    contactPerson: 'Abigail Smith',
    email: 'abigail.smith@wilmarafrica.com',
    phone: '+233 245 678 901',
    status: 'active',
    address: '12 Industrial Rd, Accra',
    notes: 'Key account. Prefers quarterly reviews.',
  },
  {
    id: 2,
    name: 'Acme Corp',
    contactPerson: 'John Doe',
    email: 'john.doe@acme.com',
    phone: '+233 501 234 567',
    status: 'inactive',
    address: '45 Market St, Kumasi',
    notes: 'On hold pending contract renewal.',
  },
])

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

const open = ref(false)
const selected = ref<ClientItem | null>(null)

function openDetails(item: ClientItem) {
  selected.value = item
  open.value = true
}
</script>

<template>
  <div class="space-y-4">
    <div class="flex flex-col gap-4">
      <!-- Search and Filters Row (match /clients design) -->
      <div class="flex flex-col md:flex-row gap-4 items-start md:items-center justify-between">
        <!-- Search Input -->
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
          <!-- Status Filter -->
          <Select v-model="statusFilter">
            <SelectTrigger class="w-[180px]" id="status">
              <SelectValue placeholder="Filter by status" />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value="all">All Status</SelectItem>
              <SelectItem value="active">Active</SelectItem>
              <SelectItem value="inactive">Inactive</SelectItem>
            </SelectContent>
          </Select>

          <!-- Clear Filters -->
          <Button v-if="searchQuery || statusFilter !== 'all'" variant="outline" @click="clearFilters">
            Clear Filters
          </Button>
        </div>
      </div>

      <!-- Results Count -->
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
            <TableHead class="w-[60px]">ID</TableHead>
            <TableHead>Client</TableHead>
            <TableHead>Contact</TableHead>
            <TableHead>Email</TableHead>
            <TableHead>Phone</TableHead>
            <TableHead>Status</TableHead>
            <TableHead class="text-right">Actions</TableHead>
          </TableRow>
        </TableHeader>
        <TableBody>
          <TableRow v-for="item in filtered" :key="item.id" class="cursor-pointer" @click="openDetails(item)">
            <TableCell>{{ item.id }}</TableCell>
            <TableCell class="font-medium">{{ item.name }}</TableCell>
            <TableCell>{{ item.contactPerson }}</TableCell>
            <TableCell>{{ item.email }}</TableCell>
            <TableCell>{{ item.phone }}</TableCell>
            <TableCell>
              <Badge :variant="item.status === 'active' ? 'default' : 'secondary'">{{ item.status }}</Badge>
            </TableCell>
            <TableCell class="text-right space-x-2">
              <Button size="sm" variant="outline" @click.stop="openDetails(item)">View</Button>
              <Button size="sm" variant="ghost">Edit</Button>
            </TableCell>
          </TableRow>
          <TableRow v-if="filtered.length === 0">
            <TableCell colspan="7" class="text-center text-muted-foreground">No clients found.</TableCell>
          </TableRow>
        </TableBody>
      </Table>
    </div>

    <!-- Offcanvas: Client Details -->
    <Sheet v-model:open="open">
      <SheetContent class="w-full sm:w-3/4 lg:w-1/2 rounded-l-lg sm:max-w-none overflow-y-auto p-6">
        <SheetHeader>
          <SheetTitle>{{ selected?.name }}</SheetTitle>
          <SheetDescription>
            Detailed information about the client
          </SheetDescription>
        </SheetHeader>
        <div class="space-y-4 py-4">
          <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
            <div>
              <div class="text-xs text-muted-foreground">Contact Person</div>
              <div class="text-sm font-medium">{{ selected?.contactPerson }}</div>
            </div>
            <div>
              <div class="text-xs text-muted-foreground">Status</div>
              <div>
                <Badge :variant="selected?.status === 'active' ? 'default' : 'secondary'">{{ selected?.status }}</Badge>
              </div>
            </div>
            <div>
              <div class="text-xs text-muted-foreground">Email</div>
              <div class="text-sm">{{ selected?.email }}</div>
            </div>
            <div>
              <div class="text-xs text-muted-foreground">Phone</div>
              <div class="text-sm">{{ selected?.phone }}</div>
            </div>
            <div v-if="selected?.address">
              <div class="text-xs text-muted-foreground">Address</div>
              <div class="text-sm">{{ selected?.address }}</div>
            </div>
          </div>

          <div v-if="selected?.notes" class="space-y-2">
            <div class="text-xs text-muted-foreground">Notes</div>
            <div class="text-sm">{{ selected?.notes }}</div>
          </div>

          <div class="pt-2 flex items-center gap-2">
            <Button variant="secondary" @click="open = false">Close</Button>
            <Button variant="outline">Edit Client</Button>
          </div>
        </div>
      </SheetContent>
    </Sheet>
  </div>
</template>
