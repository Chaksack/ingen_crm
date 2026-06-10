<script setup lang="ts">
import {
  AlertCircle,
  X,
  ChevronDown,
  Search,
} from 'lucide-vue-next'

import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardHeader, CardFooter } from '@/components/ui/card'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table'
import { Skeleton } from '@/components/ui/skeleton'
import { Input } from '@/components/ui/input'
import {
  Sheet,
  SheetContent,
  SheetDescription,
  SheetHeader,
  SheetTitle,
  SheetTrigger,
} from '@/components/ui/sheet'
import {
  DropdownMenu,
  DropdownMenuCheckboxItem,
  DropdownMenuContent,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import { ref, computed, onMounted } from 'vue'
import { useNuxtApp } from '#app'
import { Alert, AlertDescription, AlertTitle } from '~/components/ui/alert'
import { toast } from 'vue-sonner'
import { AvatarImage, Avatar, AvatarFallback } from '~/components/ui/avatar'

const financialInstitutions = ref<any[]>([])
const isLoading = ref(false)
const error = ref<string | null>(null)
const searchQuery = ref<string>('')
const statusFilter = ref<string>('all')
const currentPage = ref(1)
const itemsPerPage = ref(10)
const selectedInstitution = ref<any>(null)
const isSheetOpen = ref(false)

// Column visibility
const columnVisibility = ref({
  id: true,
  avatar: true,
  institutionName: true,
  email: true,
  status: true,
  actions: true,
})

// Filtered financial institutions based on search and status
const filteredFinancialInstitutions = computed(() => {
  let filtered = financialInstitutions.value

  // Search filter
  if (searchQuery.value) {
    filtered = filtered.filter(financialInstitution =>
      Object.values(financialInstitution).some(value =>
        String(value).toLowerCase().includes(searchQuery.value.toLowerCase())
      )
    )
  }

  // Status filter
  if (statusFilter.value !== 'all') {
    filtered = filtered.filter(financialInstitution => financialInstitution.status === statusFilter.value)
  }

  return filtered
})

// Pagination
const totalPages = computed(() => Math.ceil(filteredFinancialInstitutions.value.length / itemsPerPage.value))

const paginatedFinancialInstitutions = computed(() => {
  const start = (currentPage.value - 1) * itemsPerPage.value
  const end = start + itemsPerPage.value
  return filteredFinancialInstitutions.value.slice(start, end)
})

const goToPage = (page: number) => {
  if (page >= 1 && page <= totalPages.value) {
    currentPage.value = page
  }
}

const nextPage = () => {
  if (currentPage.value < totalPages.value) {
    currentPage.value++
  }
}

const prevPage = () => {
  if (currentPage.value > 1) {
    currentPage.value--
  }
}

// Reset to page 1 when filters change
const resetPage = () => {
  currentPage.value = 1
}

const fetchFinancialInstitutions = async () => {
  isLoading.value = true
  error.value = null

  try {
    const nuxtApp = useNuxtApp()
    const response = await nuxtApp.$axios.get('http://localhost:8080/api/financial-institutions', {
      withCredentials: true,
    })
    financialInstitutions.value = response.data
  } catch (err) {
    error.value = 'Failed to load financial institution data.'
    toast.error('Error fetching financial institutions: ' + (err as Error).message)
  } finally {
    isLoading.value = false
  }
}

const getBadgeVariant = (status: string) => {
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

const closeAlert = () => {
  error.value = null
}

const viewInstitution = (institution: any) => {
  selectedInstitution.value = institution
  isSheetOpen.value = true
}

onMounted(() => {
  fetchFinancialInstitutions()
})
</script>

<template>
  <div>
    <main class="@container/main flex flex-1 flex-col gap-4 md:gap-8">
      <Alert
        v-if="error"
        variant="default"
        class="alert bg-red-800 text-white relative"
      >
        <AlertCircle class="w-8 h-8" />
        <AlertTitle class="mx-4 font-bold">Error:</AlertTitle>
        <AlertDescription class="mx-4 text-white">{{ error }}</AlertDescription>
        
        <Button
          variant="ghost"
          size="icon"
          class="absolute top-2 right-2 h-6 w-6 text-white hover:bg-red-700"
          @click="closeAlert"
        >
          <X class="h-4 w-4" />
        </Button>
      </Alert>
      
      <Card>
        <CardHeader>
          <div class="flex flex-col gap-4">
            <!-- Search and Filters Row -->
            <div class="flex flex-col md:flex-row gap-4 items-start md:items-center justify-between">
              <!-- Search Input -->
              <div class="relative flex-1 max-w-sm">
                <Search class="absolute left-2 top-2.5 h-4 w-4 text-muted-foreground" />
                <Input
                  v-model="searchQuery"
                  placeholder="Search financial institutions..."
                  class="pl-8"
                  @input="resetPage"
                />
              </div>

              <div class="flex gap-2 flex-wrap">
                <!-- Status Filter -->
                <Select v-model="statusFilter" @update:model-value="resetPage">
                  <SelectTrigger class="w-[180px]">
                    <SelectValue placeholder="Filter by status" />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="all">All Status</SelectItem>
                    <SelectItem value="active">Active</SelectItem>
                    <SelectItem value="inactive">Inactive</SelectItem>
                    <SelectItem value="pending">Pending</SelectItem>
                  </SelectContent>
                </Select>

                <!-- Column Visibility -->
                <DropdownMenu>
                  <DropdownMenuTrigger as-child>
                    <Button variant="outline" class="ml-auto">
                      Columns <ChevronDown class="ml-2 h-4 w-4" />
                    </Button>
                  </DropdownMenuTrigger>
                  <DropdownMenuContent align="end">
                    <DropdownMenuCheckboxItem v-model:checked="columnVisibility.id">
                      ID
                    </DropdownMenuCheckboxItem>
                    <DropdownMenuCheckboxItem v-model:checked="columnVisibility.avatar">
                      Avatar
                    </DropdownMenuCheckboxItem>
                    <DropdownMenuCheckboxItem v-model:checked="columnVisibility.institutionName">
                      Institution Name
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

            <!-- Results Count -->
            <div class="text-sm text-muted-foreground">
              Showing {{ paginatedFinancialInstitutions.length }} of {{ filteredFinancialInstitutions.length }} results
              <span v-if="searchQuery || statusFilter !== 'all'">
                (filtered from {{ financialInstitutions.length }} total)
              </span>
            </div>
          </div>
        </CardHeader>

        <CardContent>
          <div class="overflow-x-auto rounded-md border">
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead v-if="columnVisibility.id">ID</TableHead>
                  <TableHead v-if="columnVisibility.avatar">Avatar</TableHead>
                  <TableHead v-if="columnVisibility.institutionName">Institution Name</TableHead>
                  <TableHead v-if="columnVisibility.email">Email</TableHead>
                  <TableHead v-if="columnVisibility.status">Status</TableHead>
                  <TableHead v-if="columnVisibility.actions">Actions</TableHead>
                </TableRow>
              </TableHeader>

              <!-- Loading State -->
              <TableBody v-if="isLoading">
                <TableRow v-for="i in itemsPerPage" :key="`skeleton-${i}`">
                  <TableCell v-if="columnVisibility.id"><Skeleton class="h-4 w-8" /></TableCell>
                  <TableCell v-if="columnVisibility.avatar"><Skeleton class="h-10 w-10 rounded-full" /></TableCell>
                  <TableCell v-if="columnVisibility.institutionName"><Skeleton class="h-4 w-24" /></TableCell>
                  <TableCell v-if="columnVisibility.email"><Skeleton class="h-4 w-40" /></TableCell>
                  <TableCell v-if="columnVisibility.status"><Skeleton class="h-6 w-16 rounded-full" /></TableCell>
                  <TableCell v-if="columnVisibility.actions"><Skeleton class="h-8 w-16 rounded-lg" /></TableCell>
                </TableRow>
              </TableBody>

              <!-- Data State -->
              <TableBody v-else-if="paginatedFinancialInstitutions && paginatedFinancialInstitutions.length > 0">
                <TableRow v-for="financialInstitution in paginatedFinancialInstitutions" :key="financialInstitution.ID">
                  <TableCell v-if="columnVisibility.id">{{ financialInstitution.ID || 'N/A' }}</TableCell>
                  <TableCell v-if="columnVisibility.avatar">
                    <Avatar class="relative overflow-visible">
                      <AvatarImage class="rounded-full" :src="financialInstitution.avatar || ''" alt="Financial Institution Avatar" />
                      <AvatarFallback class="text-white">
                        {{ financialInstitution.institution_name ? financialInstitution.institution_name.substring(0, 2).toUpperCase() : 'FI' }}
                      </AvatarFallback>
                      <span
                        v-if="financialInstitution"
                        :class="financialInstitution.status === 'active' ? 'bg-green-500' : 'bg-red-500'"
                        class="absolute bottom-[-4px] right-[-4px] w-3.5 h-3.5 rounded-full border-2 border-white"
                      />
                    </Avatar>
                  </TableCell>
                  <TableCell v-if="columnVisibility.institutionName">{{ financialInstitution.institution_name || 'N/A' }}</TableCell>
                  <TableCell v-if="columnVisibility.email">{{ financialInstitution.email || 'N/A' }}</TableCell>
                  <TableCell v-if="columnVisibility.status">
                    <Badge :variant="getBadgeVariant(financialInstitution.status)">
                      {{ financialInstitution.status || 'N/A' }}
                    </Badge>
                  </TableCell>
                  <TableCell v-if="columnVisibility.actions">
                    <Button size="sm" variant="default" @click="viewInstitution(financialInstitution)">
                      View
                    </Button>
                  </TableCell>
                </TableRow>
              </TableBody>
              
              <!-- Empty State -->
              <TableBody v-else>
                <TableRow>
                  <TableCell :colspan="6" class="h-64 text-center">
                    <div class="flex flex-col items-center justify-center py-12">
                      <div class="rounded-full bg-gray-100 p-4 mb-4">
                        <AlertCircle class="w-12 h-12 text-gray-400" />
                      </div>
                      <h3 class="text-lg font-semibold mb-2">
                        No Institution Data Available
                      </h3>
                      <p class="text-sm max-w-sm">
                        {{ searchQuery || statusFilter !== 'all' 
                          ? 'No institutions match your search criteria. Try adjusting your filters.' 
                          : 'No institution data available. Get started by creating your first institution.' 
                        }}
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
          <!-- Items per page selector -->
          <div class="flex items-center gap-2">
            <span class="text-sm text-muted-foreground">Rows per page:</span>
            <Select v-model="itemsPerPage" @update:model-value="resetPage">
              <SelectTrigger class="w-[70px]">
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
                <SelectItem :value="5">5</SelectItem>
                <SelectItem :value="10">10</SelectItem>
                <SelectItem :value="20">20</SelectItem>
                <SelectItem :value="50">50</SelectItem>
              </SelectContent>
            </Select>
          </div>

          <!-- Pagination Controls -->
          <div class="flex items-center gap-2">
            <span class="text-sm text-muted-foreground">
              Page {{ currentPage }} of {{ totalPages || 1 }}
            </span>
            <div class="flex gap-1">
              <Button
                variant="outline"
                size="sm"
                :disabled="currentPage === 1"
                @click="prevPage"
              >
                Previous
              </Button>
              
              <!-- Page Numbers -->
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

              <Button
                variant="outline"
                size="sm"
                :disabled="currentPage === totalPages"
                @click="nextPage"
              >
                Next
              </Button>
            </div>
          </div>
        </CardFooter>
      </Card>
    </main>

    <!-- Sheet for viewing institution details -->
    <Sheet v-model:open="isSheetOpen">
      <SheetContent side="right" class="w-full sm:max-w-2xl overflow-y-auto">
        <SheetHeader>
          <SheetTitle class="flex items-center gap-3">
            <Avatar class="h-12 w-12">
              <AvatarImage :src="selectedInstitution?.avatar || ''" alt="Institution Avatar" />
              <AvatarFallback>
                {{ selectedInstitution?.institution_name ? selectedInstitution.institution_name.substring(0, 2).toUpperCase() : 'FI' }}
              </AvatarFallback>
            </Avatar>
            <div>
              <div class="text-xl font-semibold">{{ selectedInstitution?.institution_name || 'N/A' }}</div>
              <div class="text-sm text-muted-foreground font-normal">ID: {{ selectedInstitution?.ID || 'N/A' }}</div>
            </div>
          </SheetTitle>
          <SheetDescription>
            View and manage financial institution details
          </SheetDescription>
        </SheetHeader>

        <div v-if="selectedInstitution" class="mt-6 space-y-6">
          <!-- Status Badge -->
          <div class="flex items-center gap-2">
            <span class="text-sm font-medium">Status:</span>
            <Badge :variant="getBadgeVariant(selectedInstitution.status)">
              {{ selectedInstitution.status || 'N/A' }}
            </Badge>
          </div>

          <!-- Contact Information -->
          <Card>
            <CardHeader>
              <h3 class="text-lg font-semibold">Contact Information</h3>
            </CardHeader>
            <CardContent class="space-y-3">
              <div class="grid grid-cols-3 gap-2">
                <span class="text-sm text-muted-foreground">Email:</span>
                <span class="col-span-2 text-sm font-medium">{{ selectedInstitution.email || 'N/A' }}</span>
              </div>
              <div class="grid grid-cols-3 gap-2">
                <span class="text-sm text-muted-foreground">Phone:</span>
                <span class="col-span-2 text-sm font-medium">{{ selectedInstitution.phone || 'N/A' }}</span>
              </div>
              <div class="grid grid-cols-3 gap-2">
                <span class="text-sm text-muted-foreground">Address:</span>
                <span class="col-span-2 text-sm font-medium">{{ selectedInstitution.address || 'N/A' }}</span>
              </div>
              <div class="grid grid-cols-3 gap-2">
                <span class="text-sm text-muted-foreground">Website:</span>
                <span class="col-span-2 text-sm font-medium">
                  <a v-if="selectedInstitution.website" :href="selectedInstitution.website" target="_blank" class="text-primary hover:underline">
                    {{ selectedInstitution.website }}
                  </a>
                  <span v-else>N/A</span>
                </span>
              </div>
            </CardContent>
          </Card>

          <!-- Institution Details -->
          <Card>
            <CardHeader>
              <h3 class="text-lg font-semibold">Institution Details</h3>
            </CardHeader>
            <CardContent class="space-y-3">
              <div class="grid grid-cols-3 gap-2">
                <span class="text-sm text-muted-foreground">Type:</span>
                <span class="col-span-2 text-sm font-medium">{{ selectedInstitution.institution_type || 'N/A' }}</span>
              </div>
              <div class="grid grid-cols-3 gap-2">
                <span class="text-sm text-muted-foreground">Registration Number:</span>
                <span class="col-span-2 text-sm font-medium">{{ selectedInstitution.registration_number || 'N/A' }}</span>
              </div>
              <div class="grid grid-cols-3 gap-2">
                <span class="text-sm text-muted-foreground">Founded:</span>
                <span class="col-span-2 text-sm font-medium">{{ selectedInstitution.founded_date || 'N/A' }}</span>
              </div>
              <div class="grid grid-cols-3 gap-2">
                <span class="text-sm text-muted-foreground">Country:</span>
                <span class="col-span-2 text-sm font-medium">{{ selectedInstitution.country || 'N/A' }}</span>
              </div>
            </CardContent>
          </Card>

          <!-- Additional Information -->
          <Card>
            <CardHeader>
              <h3 class="text-lg font-semibold">Additional Information</h3>
            </CardHeader>
            <CardContent class="space-y-3">
              <div class="grid grid-cols-3 gap-2">
                <span class="text-sm text-muted-foreground">Services:</span>
                <span class="col-span-2 text-sm font-medium">{{ selectedInstitution.services || 'N/A' }}</span>
              </div>
              <div class="grid grid-cols-3 gap-2">
                <span class="text-sm text-muted-foreground">Description:</span>
                <span class="col-span-2 text-sm font-medium">{{ selectedInstitution.description || 'N/A' }}</span>
              </div>
            </CardContent>
          </Card>

          <!-- Action Buttons -->
          <div class="flex gap-2 pt-4">
            <Button class="flex-1">Edit Institution</Button>
            <Button variant="outline" class="flex-1" @click="isSheetOpen = false">Close</Button>
          </div>
        </div>
      </SheetContent>
    </Sheet>
  </div>
</template>