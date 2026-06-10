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

const businesses = ref<any[]>([])
const isLoading = ref(false)
const error = ref<string | null>(null)
const searchQuery = ref<string>('')
const statusFilter = ref<string>('all')
const currentPage = ref(1)
const itemsPerPage = ref(10)

// Column visibility
const columnVisibility = ref({
  id: true,
  avatar: true,
  companyName: true,
  capital: true,
  foundedYear: true,
  email: true,
  phoneNumber: true,
  status: true,
  creditScore: true,
  actions: true,
})

// Filtered businesses based on search and status
const filteredBusinesses = computed(() => {
  let filtered = businesses.value

  // Search filter
  if (searchQuery.value) {
    filtered = filtered.filter(business =>
      Object.values(business).some(value =>
        String(value).toLowerCase().includes(searchQuery.value.toLowerCase())
      )
    )
  }

  // Status filter
  if (statusFilter.value !== 'all') {
    filtered = filtered.filter(customer => customer.status === statusFilter.value)
  }

  return filtered
})

// Pagination
const totalPages = computed(() => Math.ceil(filteredBusinesses.value.length / itemsPerPage.value))

const paginatedBusinesses = computed(() => {
  const start = (currentPage.value - 1) * itemsPerPage.value
  const end = start + itemsPerPage.value
  return filteredBusinesses.value.slice(start, end)
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

const fetchBusiness = async () => {
  isLoading.value = true
  error.value = null

  try {
    const nuxtApp = useNuxtApp()
    const response = await nuxtApp.$axios.get('http://localhost:8080/api/businesses', {
      withCredentials: true,
    })
    businesses.value = response.data
  } catch (err) {
    error.value = 'Failed to load business data.'
    toast.error('Error fetching business: ' + (err as Error).message)
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

onMounted(() => {
  fetchBusiness()
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
                  placeholder="Search businesses..."
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
                    <DropdownMenuCheckboxItem v-model:checked="columnVisibility.companyName">
                      Company Name
                    </DropdownMenuCheckboxItem>
                    <DropdownMenuCheckboxItem v-model:checked="columnVisibility.capital">
                      Capital
                    </DropdownMenuCheckboxItem>
                    <DropdownMenuCheckboxItem v-model:checked="columnVisibility.foundedYear">
                      Founded Year
                    </DropdownMenuCheckboxItem>
                    <DropdownMenuCheckboxItem v-model:checked="columnVisibility.email">
                      Email
                    </DropdownMenuCheckboxItem>
                    <DropdownMenuCheckboxItem v-model:checked="columnVisibility.phoneNumber">
                      Phone Number
                    </DropdownMenuCheckboxItem>
                    <DropdownMenuCheckboxItem v-model:checked="columnVisibility.status">
                      Status
                    </DropdownMenuCheckboxItem>
                    <DropdownMenuCheckboxItem v-model:checked="columnVisibility.creditScore">
                      Credit Score
                    </DropdownMenuCheckboxItem>
                  </DropdownMenuContent>
                </DropdownMenu>
              </div>
            </div>

            <!-- Results Count -->
            <div class="text-sm text-muted-foreground">
              Showing {{ paginatedBusinesses.length }} of {{ filteredBusinesses.length }} results
              <span v-if="searchQuery || statusFilter !== 'all'">
                (filtered from {{ businesses.length }} total)
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
                  <TableHead v-if="columnVisibility.companyName">Company Name</TableHead>
                  <TableHead v-if="columnVisibility.email">Email</TableHead>
                  <TableHead v-if="columnVisibility.phoneNumber">Phone Number</TableHead>
                  <TableHead v-if="columnVisibility.capital">Capital</TableHead>
                  <TableHead v-if="columnVisibility.foundedYear">Founded Year</TableHead>
                  <TableHead v-if="columnVisibility.status">Status</TableHead>
                  <TableHead v-if="columnVisibility.creditScore">Credit Score</TableHead>
                  <TableHead v-if="columnVisibility.actions">Actions</TableHead>
                </TableRow>
              </TableHeader>

              <!-- Loading State -->
              <TableBody v-if="isLoading">
                <TableRow v-for="i in itemsPerPage" :key="`skeleton-${i}`">
                  <TableCell v-if="columnVisibility.id"><Skeleton class="h-4 w-8" /></TableCell>
                  <TableCell v-if="columnVisibility.avatar"><Skeleton class="h-10 w-10 rounded-full" /></TableCell>
                  <TableCell v-if="columnVisibility.companyName"><Skeleton class="h-4 w-24" /></TableCell>
                  <TableCell v-if="columnVisibility.capital"><Skeleton class="h-4 w-24" /></TableCell>
                  <TableCell v-if="columnVisibility.foundedYear"><Skeleton class="h-4 w-8" /></TableCell>
                  <TableCell v-if="columnVisibility.email"><Skeleton class="h-4 w-40" /></TableCell>
                  <TableCell v-if="columnVisibility.phoneNumber"><Skeleton class="h-4 w-28" /></TableCell>
                  <TableCell v-if="columnVisibility.status"><Skeleton class="h-6 w-16 rounded-full" /></TableCell>
                  <TableCell v-if="columnVisibility.creditScore"><Skeleton class="h-4 w-12" /></TableCell>
                  <TableCell v-if="columnVisibility.actions"><Skeleton class="h-8 w-16 rounded-lg" /></TableCell>
                </TableRow>
              </TableBody>

              <!-- Data State -->
              <TableBody v-else-if="paginatedBusinesses && paginatedBusinesses.length > 0">
                <TableRow v-for="business in paginatedBusinesses" :key="business.ID">
                  <TableCell v-if="columnVisibility.id">{{ business.ID || 'N/A' }}</TableCell>
                  <TableCell v-if="columnVisibility.avatar">
                    <Avatar class="relative overflow-visible">
                      <AvatarImage class="rounded-full" :src="business.avatar || ''" alt="Business Avatar" />
                      <AvatarFallback class="text-white">
                        {{ business.companyName ? `${business.companyName[0]}` : '' }}
                      </AvatarFallback>
                      <span
                        v-if="business"
                        :class="business.status === 'active' ? 'bg-green-500' : 'bg-red-500'"
                        class="absolute bottom-[-4px] right-[-4px] w-3.5 h-3.5 rounded-full border-2 border-white"
                      />
                    </Avatar>
                  </TableCell>
                  <TableCell v-if="columnVisibility.companyName">{{ business.companyName || 'N/A' }}</TableCell>
                  <TableCell v-if="columnVisibility.capital">{{ business.capital || 'N/A' }}</TableCell>
                  <TableCell v-if="columnVisibility.foundedYear">{{ business.foundedYear || 'N/A' }}</TableCell>
                  <TableCell v-if="columnVisibility.email">{{ business.email || 'N/A' }}</TableCell>
                  <TableCell v-if="columnVisibility.phoneNumber">{{ business.phoneNumber || 'N/A' }}</TableCell>
                  <TableCell v-if="columnVisibility.status">
                    <Badge :variant="getBadgeVariant(business.status)">
                      {{ business.status || 'N/A' }}
                    </Badge>
                  </TableCell>
                  <TableCell v-if="columnVisibility.creditScore" class="font-bold">
                    {{ business.creditScore || 'N/A' }}
                  </TableCell>
                  <TableCell v-if="columnVisibility.actions">
                    <NuxtLink :to="`/company/view/${business.ID}`">
                      <Button size="sm" variant="default">
                        View
                      </Button>
                    </NuxtLink>
                  </TableCell>
                </TableRow>
              </TableBody>
              
              <!-- Empty State -->
              <TableBody v-else>
                <TableRow>
                  <TableCell :colspan="10" class="h-64 text-center">
                    <div class="flex flex-col items-center justify-center py-12">
                      <div class="rounded-full bg-gray-100 p-4 mb-4">
                        <AlertCircle class="w-12 h-12 text-gray-400" />
                      </div>
                      <h3 class="text-lg font-semibold mb-2">
                        No Business Data Available
                      </h3>
                      <p class="text-sm max-w-sm">
                        {{ searchQuery || statusFilter !== 'all' 
                          ? 'No businesses match your search criteria. Try adjusting your filters.' 
                          : 'No business data available. Get started by creating your first business.' 
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
  </div>
</template>