<script setup lang="ts">
import {
  AlertCircle,
  ChevronDown,
  Search,
} from 'lucide-vue-next'

import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardFooter, CardHeader } from '@/components/ui/card'
import {
  DropdownMenu,
  DropdownMenuCheckboxItem,
  DropdownMenuContent,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu'
import { Input } from '@/components/ui/input'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import {
  Sheet,
  SheetContent,
  SheetDescription,
  SheetHeader,
  SheetTitle,
} from '@/components/ui/sheet'
import { Skeleton } from '@/components/ui/skeleton'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table'
import { Alert, AlertDescription, AlertTitle } from '~/components/ui/alert'
import { Avatar, AvatarFallback, AvatarImage } from '~/components/ui/avatar'

interface Vendor {
  id: string
  vendorName: string
  email: string
  phone?: string
  address?: string
  website?: string
  avatar?: string
  vendorType?: string
  registrationNumber?: string
  foundedDate?: string
  country?: string
  services?: string
  description?: string
  status: 'active' | 'inactive' | 'pending'
}

const { data: vendors, pending: isLoading, error, refresh } = useFetch<Vendor[]>('/api/vendors')

const searchQuery = ref('')
const statusFilter = ref('all')
const currentPage = ref(1)
const itemsPerPage = ref(10)
const selectedVendor = ref<Vendor | null>(null)
const isSheetOpen = ref(false)

const columnVisibility = ref({
  avatar: true,
  vendorName: true,
  email: true,
  status: true,
  actions: true,
})

const filteredVendors = computed(() => {
  let filtered = vendors.value ?? []

  if (searchQuery.value) {
    filtered = filtered.filter(vendor =>
      Object.values(vendor).some(value =>
        String(value).toLowerCase().includes(searchQuery.value.toLowerCase()),
      ),
    )
  }

  if (statusFilter.value !== 'all')
    filtered = filtered.filter(vendor => vendor.status === statusFilter.value)

  return filtered
})

const totalPages = computed(() => Math.ceil(filteredVendors.value.length / itemsPerPage.value))

const paginatedVendors = computed(() => {
  const start = (currentPage.value - 1) * itemsPerPage.value
  return filteredVendors.value.slice(start, start + itemsPerPage.value)
})

function goToPage(page: number) {
  if (page >= 1 && page <= totalPages.value)
    currentPage.value = page
}

function nextPage() {
  if (currentPage.value < totalPages.value)
    currentPage.value++
}

function prevPage() {
  if (currentPage.value > 1)
    currentPage.value--
}

function resetPage() {
  currentPage.value = 1
}

function getBadgeVariant(status: string) {
  switch (status) {
    case 'active':
      return 'default'
    case 'inactive':
      return 'destructive'
    case 'pending':
      return 'secondary'
    default:
      return 'secondary'
  }
}

function viewVendor(vendor: Vendor) {
  selectedVendor.value = vendor
  isSheetOpen.value = true
}

defineExpose({ refresh })
</script>

<template>
  <div>
    <main class="@container/main flex flex-1 flex-col gap-4 md:gap-8">
      <Alert
        v-if="error"
        variant="destructive"
      >
        <AlertCircle class="w-4 h-4" />
        <AlertTitle>Error</AlertTitle>
        <AlertDescription>Failed to load vendors.</AlertDescription>
      </Alert>

      <Card>
        <CardHeader>
          <div class="flex flex-col gap-4">
            <div class="flex flex-col md:flex-row gap-4 items-start md:items-center justify-between">
              <div class="relative flex-1 max-w-sm">
                <Search class="absolute left-2 top-2.5 h-4 w-4 text-muted-foreground" />
                <Input
                  v-model="searchQuery"
                  placeholder="Search vendors..."
                  class="pl-8"
                  @input="resetPage"
                />
              </div>

              <div class="flex gap-2 flex-wrap">
                <Select v-model="statusFilter" @update:model-value="resetPage">
                  <SelectTrigger class="w-[180px]">
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
                    <SelectItem value="pending">
                      Pending
                    </SelectItem>
                  </SelectContent>
                </Select>

                <DropdownMenu>
                  <DropdownMenuTrigger as-child>
                    <Button variant="outline" class="ml-auto">
                      Columns <ChevronDown class="ml-2 h-4 w-4" />
                    </Button>
                  </DropdownMenuTrigger>
                  <DropdownMenuContent align="end">
                    <DropdownMenuCheckboxItem v-model:checked="columnVisibility.avatar">
                      Avatar
                    </DropdownMenuCheckboxItem>
                    <DropdownMenuCheckboxItem v-model:checked="columnVisibility.vendorName">
                      Vendor Name
                    </DropdownMenuCheckboxItem>
                    <DropdownMenuCheckboxItem v-model:checked="columnVisibility.email">
                      Email
                    </DropdownMenuCheckboxItem>
                    <DropdownMenuCheckboxItem v-model:checked="columnVisibility.status">
                      Status
                    </DropdownMenuCheckboxItem>
                  </DropdownMenuContent>
                </DropdownMenu>
              </div>
            </div>

            <div class="text-sm text-muted-foreground">
              Showing {{ paginatedVendors.length }} of {{ filteredVendors.length }} results
              <span v-if="searchQuery || statusFilter !== 'all'">
                (filtered from {{ vendors?.length ?? 0 }} total)
              </span>
            </div>
          </div>
        </CardHeader>

        <CardContent>
          <div class="overflow-x-auto rounded-md border">
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead v-if="columnVisibility.avatar">
                    Avatar
                  </TableHead>
                  <TableHead v-if="columnVisibility.vendorName">
                    Vendor Name
                  </TableHead>
                  <TableHead v-if="columnVisibility.email">
                    Email
                  </TableHead>
                  <TableHead v-if="columnVisibility.status">
                    Status
                  </TableHead>
                  <TableHead v-if="columnVisibility.actions">
                    Actions
                  </TableHead>
                </TableRow>
              </TableHeader>

              <TableBody v-if="isLoading">
                <TableRow v-for="i in itemsPerPage" :key="`skeleton-${i}`">
                  <TableCell v-if="columnVisibility.avatar">
                    <Skeleton class="h-10 w-10 rounded-full" />
                  </TableCell>
                  <TableCell v-if="columnVisibility.vendorName">
                    <Skeleton class="h-4 w-24" />
                  </TableCell>
                  <TableCell v-if="columnVisibility.email">
                    <Skeleton class="h-4 w-40" />
                  </TableCell>
                  <TableCell v-if="columnVisibility.status">
                    <Skeleton class="h-6 w-16 rounded-full" />
                  </TableCell>
                  <TableCell v-if="columnVisibility.actions">
                    <Skeleton class="h-8 w-16 rounded-lg" />
                  </TableCell>
                </TableRow>
              </TableBody>

              <TableBody v-else-if="paginatedVendors.length > 0">
                <TableRow v-for="vendor in paginatedVendors" :key="vendor.id">
                  <TableCell v-if="columnVisibility.avatar">
                    <Avatar class="relative overflow-visible">
                      <AvatarImage class="rounded-full" :src="vendor.avatar || ''" alt="Vendor Avatar" />
                      <AvatarFallback class="text-white">
                        {{ vendor.vendorName.substring(0, 2).toUpperCase() }}
                      </AvatarFallback>
                      <span
                        :class="vendor.status === 'active' ? 'bg-green-500' : 'bg-red-500'"
                        class="absolute bottom-[-4px] right-[-4px] w-3.5 h-3.5 rounded-full border-2 border-white"
                      />
                    </Avatar>
                  </TableCell>
                  <TableCell v-if="columnVisibility.vendorName">
                    {{ vendor.vendorName }}
                  </TableCell>
                  <TableCell v-if="columnVisibility.email">
                    {{ vendor.email }}
                  </TableCell>
                  <TableCell v-if="columnVisibility.status">
                    <Badge :variant="getBadgeVariant(vendor.status)">
                      {{ vendor.status }}
                    </Badge>
                  </TableCell>
                  <TableCell v-if="columnVisibility.actions">
                    <Button size="sm" variant="default" @click="viewVendor(vendor)">
                      View
                    </Button>
                  </TableCell>
                </TableRow>
              </TableBody>

              <TableBody v-else>
                <TableRow>
                  <TableCell :colspan="5" class="h-64 text-center">
                    <div class="flex flex-col items-center justify-center py-12">
                      <div class="rounded-full bg-muted p-4 mb-4">
                        <AlertCircle class="w-12 h-12 text-muted-foreground" />
                      </div>
                      <h3 class="text-lg font-semibold mb-2">
                        No Vendor Data Available
                      </h3>
                      <p class="text-sm max-w-sm text-muted-foreground">
                        {{ searchQuery || statusFilter !== 'all'
                          ? 'No vendors match your search criteria. Try adjusting your filters.'
                          : 'No vendors yet. Get started by creating your first vendor.' }}
                      </p>
                      <Button
                        v-if="searchQuery || statusFilter !== 'all'"
                        variant="outline"
                        class="mt-4"
                        @click="searchQuery = ''; statusFilter = 'all'"
                      >
                        Clear Filters
                      </Button>
                    </div>
                  </TableCell>
                </TableRow>
              </TableBody>
            </Table>
          </div>
        </CardContent>

        <CardFooter class="flex items-center justify-between">
          <div class="flex items-center gap-2">
            <span class="text-sm text-muted-foreground">Rows per page:</span>
            <Select v-model="itemsPerPage" @update:model-value="resetPage">
              <SelectTrigger class="w-[70px]">
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
                <SelectItem :value="5">
                  5
                </SelectItem>
                <SelectItem :value="10">
                  10
                </SelectItem>
                <SelectItem :value="20">
                  20
                </SelectItem>
                <SelectItem :value="50">
                  50
                </SelectItem>
              </SelectContent>
            </Select>
          </div>

          <div class="flex items-center gap-2">
            <span class="text-sm text-muted-foreground">
              Page {{ currentPage }} of {{ totalPages || 1 }}
            </span>
            <div class="flex gap-1">
              <Button variant="outline" size="sm" :disabled="currentPage === 1" @click="prevPage">
                Previous
              </Button>
              <template v-for="page in totalPages" :key="page">
                <Button
                  v-if="page === 1 || page === totalPages || (page >= currentPage - 1 && page <= currentPage + 1)"
                  variant="outline"
                  size="sm"
                  :class="{ 'bg-primary text-primary-foreground': page === currentPage }"
                  @click="goToPage(page)"
                >
                  {{ page }}
                </Button>
                <span
                  v-else-if="page === currentPage - 2 || page === currentPage + 2"
                  class="flex items-center px-2"
                >
                  ...
                </span>
              </template>
              <Button variant="outline" size="sm" :disabled="currentPage === totalPages" @click="nextPage">
                Next
              </Button>
            </div>
          </div>
        </CardFooter>
      </Card>
    </main>

    <Sheet v-model:open="isSheetOpen">
      <SheetContent side="right" class="w-full sm:max-w-2xl overflow-y-auto">
        <SheetHeader>
          <SheetTitle class="flex items-center gap-3">
            <Avatar class="h-12 w-12">
              <AvatarImage :src="selectedVendor?.avatar || ''" alt="Vendor Avatar" />
              <AvatarFallback>
                {{ selectedVendor?.vendorName ? selectedVendor.vendorName.substring(0, 2).toUpperCase() : 'VN' }}
              </AvatarFallback>
            </Avatar>
            <div>
              <div class="text-xl font-semibold">
                {{ selectedVendor?.vendorName }}
              </div>
              <div class="text-sm text-muted-foreground font-normal">
                {{ selectedVendor?.email }}
              </div>
            </div>
          </SheetTitle>
          <SheetDescription>
            View vendor details
          </SheetDescription>
        </SheetHeader>

        <div v-if="selectedVendor" class="mt-6 space-y-6 px-4">
          <div class="flex items-center gap-2">
            <span class="text-sm font-medium">Status:</span>
            <Badge :variant="getBadgeVariant(selectedVendor.status)">
              {{ selectedVendor.status }}
            </Badge>
          </div>

          <Card>
            <CardHeader>
              <h3 class="text-lg font-semibold">
                Contact Information
              </h3>
            </CardHeader>
            <CardContent class="space-y-3">
              <div class="grid grid-cols-3 gap-2">
                <span class="text-sm text-muted-foreground">Phone:</span>
                <span class="col-span-2 text-sm font-medium">{{ selectedVendor.phone || 'N/A' }}</span>
              </div>
              <div class="grid grid-cols-3 gap-2">
                <span class="text-sm text-muted-foreground">Address:</span>
                <span class="col-span-2 text-sm font-medium">{{ selectedVendor.address || 'N/A' }}</span>
              </div>
              <div class="grid grid-cols-3 gap-2">
                <span class="text-sm text-muted-foreground">Website:</span>
                <span class="col-span-2 text-sm font-medium">
                  <a v-if="selectedVendor.website" :href="selectedVendor.website" target="_blank" class="text-primary hover:underline">
                    {{ selectedVendor.website }}
                  </a>
                  <span v-else>N/A</span>
                </span>
              </div>
            </CardContent>
          </Card>

          <Card>
            <CardHeader>
              <h3 class="text-lg font-semibold">
                Vendor Details
              </h3>
            </CardHeader>
            <CardContent class="space-y-3">
              <div class="grid grid-cols-3 gap-2">
                <span class="text-sm text-muted-foreground">Type:</span>
                <span class="col-span-2 text-sm font-medium">{{ selectedVendor.vendorType || 'N/A' }}</span>
              </div>
              <div class="grid grid-cols-3 gap-2">
                <span class="text-sm text-muted-foreground">Registration Number:</span>
                <span class="col-span-2 text-sm font-medium">{{ selectedVendor.registrationNumber || 'N/A' }}</span>
              </div>
              <div class="grid grid-cols-3 gap-2">
                <span class="text-sm text-muted-foreground">Founded:</span>
                <span class="col-span-2 text-sm font-medium">{{ selectedVendor.foundedDate || 'N/A' }}</span>
              </div>
              <div class="grid grid-cols-3 gap-2">
                <span class="text-sm text-muted-foreground">Country:</span>
                <span class="col-span-2 text-sm font-medium">{{ selectedVendor.country || 'N/A' }}</span>
              </div>
            </CardContent>
          </Card>

          <Card>
            <CardHeader>
              <h3 class="text-lg font-semibold">
                Additional Information
              </h3>
            </CardHeader>
            <CardContent class="space-y-3">
              <div class="grid grid-cols-3 gap-2">
                <span class="text-sm text-muted-foreground">Services:</span>
                <span class="col-span-2 text-sm font-medium">{{ selectedVendor.services || 'N/A' }}</span>
              </div>
              <div class="grid grid-cols-3 gap-2">
                <span class="text-sm text-muted-foreground">Description:</span>
                <span class="col-span-2 text-sm font-medium">{{ selectedVendor.description || 'N/A' }}</span>
              </div>
            </CardContent>
          </Card>

          <div class="flex gap-2 pt-4">
            <Button variant="outline" class="flex-1" @click="isSheetOpen = false">
              Close
            </Button>
          </div>
        </div>
      </SheetContent>
    </Sheet>
  </div>
</template>
