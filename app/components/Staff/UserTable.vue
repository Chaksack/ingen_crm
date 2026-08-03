<script setup lang="ts">
import { Search } from 'lucide-vue-next'
import { computed, ref } from 'vue'
import { toast } from 'vue-sonner'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { Skeleton } from '@/components/ui/skeleton'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table'

interface StaffItem {
  id: string
  firstName: string
  lastName: string
  email: string
  phoneNumber?: string
  role: 'Staff' | 'Manager' | 'Admin'
  department?: string
  status: 'active' | 'inactive'
  loginStatus: 'pending' | 'active' | 'inactive' | null
}

const { data: staffData, pending: isLoading, refresh } = useFetch<StaffItem[]>('/api/staff')
const staff = computed(() => staffData.value ?? [])
const { user: authUser } = useAuth()
const isAdmin = computed(() => authUser.value?.role === 'admin')

const searchQuery = ref('')
const roleFilter = ref<'all' | 'Staff' | 'Manager' | 'Admin'>('all')
const statusFilter = ref<'all' | 'active' | 'inactive'>('all')

const filtered = computed(() => {
  let items = staff.value
  const q = searchQuery.value.trim().toLowerCase()
  if (q) {
    items = items.filter(i =>
      [i.firstName, i.lastName, i.email, i.role, i.status].some(v => String(v).toLowerCase().includes(q)),
    )
  }
  if (roleFilter.value !== 'all')
    items = items.filter(i => i.role === roleFilter.value)
  if (statusFilter.value !== 'all')
    items = items.filter(i => i.status === statusFilter.value)
  return items
})

function clearFilters() {
  searchQuery.value = ''
  roleFilter.value = 'all'
  statusFilter.value = 'all'
}

const open = ref(false)
const selected = ref<StaffItem | null>(null)

function openDetails(item: StaffItem) {
  selected.value = item
  open.value = true
}

function loginStatusLabel(status: StaffItem['loginStatus']) {
  switch (status) {
    case 'active': return 'Active login'
    case 'pending': return 'Invite pending'
    case 'inactive': return 'Access revoked'
    default: return 'No login'
  }
}

function loginStatusVariant(status: StaffItem['loginStatus']) {
  switch (status) {
    case 'active': return 'default'
    case 'pending': return 'secondary'
    case 'inactive': return 'destructive'
    default: return 'outline'
  }
}

const isResending = ref(false)
async function resendInvite(item: StaffItem) {
  isResending.value = true
  try {
    const result = await $fetch<{ inviteEmailSent: boolean, inviteLink: string | null }>(`/api/staff/${item.id}/resend-invite`, { method: 'POST' })
    if (result.inviteEmailSent) {
      toast.success('Invite email resent')
    }
    else {
      toast.warning('Could not send the invite email. Share this link with them directly:', {
        description: result.inviteLink ?? undefined,
        duration: 15000,
      })
    }
  }
  catch (err: any) {
    toast.error(err?.data?.statusMessage || 'Failed to resend invite')
  }
  finally {
    isResending.value = false
  }
}

const isDeleting = ref(false)
async function deleteStaff(item: StaffItem) {
  isDeleting.value = true
  try {
    await $fetch(`/api/staff/${item.id}`, { method: 'DELETE' })
    toast.success('Staff member deleted')
    open.value = false
    await refresh()
  }
  catch (err: any) {
    toast.error(err?.data?.statusMessage || 'Failed to delete staff member')
  }
  finally {
    isDeleting.value = false
  }
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
            placeholder="Search staff..."
            class="pl-8"
          />
        </div>

        <div class="flex gap-2 flex-wrap">
          <Select v-model="roleFilter">
            <SelectTrigger id="role" class="w-[180px]">
              <SelectValue placeholder="Filter by role" />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value="all">
                All Roles
              </SelectItem>
              <SelectItem value="Staff">
                Staff
              </SelectItem>
              <SelectItem value="Manager">
                Manager
              </SelectItem>
              <SelectItem value="Admin">
                Admin
              </SelectItem>
            </SelectContent>
          </Select>

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

          <Button v-if="searchQuery || roleFilter !== 'all' || statusFilter !== 'all'" variant="outline" @click="clearFilters">
            Clear Filters
          </Button>
        </div>
      </div>

      <div class="text-sm text-muted-foreground">
        Showing {{ filtered.length }} of {{ staff.length }} results
        <span v-if="searchQuery || roleFilter !== 'all' || statusFilter !== 'all'">
          (filtered)
        </span>
      </div>
    </div>

    <div class="rounded-md border">
      <Table>
        <TableHeader>
          <TableRow>
            <TableHead>Name</TableHead>
            <TableHead>Email</TableHead>
            <TableHead>Role</TableHead>
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
              <Skeleton class="h-4 w-40" />
            </TableCell>
            <TableCell>
              <Skeleton class="h-4 w-20" />
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
          <TableRow v-for="item in filtered" :key="item.id" class="cursor-pointer" @click="openDetails(item)">
            <TableCell>
              {{ item.firstName }} {{ item.lastName }}
            </TableCell>
            <TableCell>
              {{ item.email }}
            </TableCell>
            <TableCell>
              {{ item.role }}
            </TableCell>
            <TableCell>
              <Badge :variant="item.status === 'active' ? 'default' : 'secondary'">
                {{ item.status }}
              </Badge>
            </TableCell>
            <TableCell class="text-right space-x-2">
              <Button size="sm" variant="outline" @click.stop="openDetails(item)">
                View
              </Button>
            </TableCell>
          </TableRow>
          <TableRow v-if="filtered.length === 0">
            <TableCell colspan="5" class="text-center text-muted-foreground">
              No staff found.
            </TableCell>
          </TableRow>
        </TableBody>
      </Table>
    </div>

    <Sheet v-model:open="open">
      <SheetContent class="w-full sm:w-3/4 lg:w-1/2 rounded-l-lg sm:max-w-none overflow-y-auto p-6">
        <SheetHeader>
          <SheetTitle>{{ selected ? `${selected.firstName} ${selected.lastName}` : '' }}</SheetTitle>
          <SheetDescription>
            Detailed information about the staff member
          </SheetDescription>
        </SheetHeader>
        <div class="space-y-4 py-4">
          <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
            <div>
              <div class="text-xs text-muted-foreground">
                Role
              </div>
              <div class="text-sm font-medium">
                {{ selected?.role }}
              </div>
            </div>
            <div>
              <div class="text-xs text-muted-foreground">
                Status
              </div>
              <div>
                <Badge :variant="selected?.status === 'active' ? 'default' : 'secondary'">
                  {{ selected?.status }}
                </Badge>
              </div>
            </div>
            <div>
              <div class="text-xs text-muted-foreground">
                Email
              </div>
              <div class="text-sm">
                {{ selected?.email }}
              </div>
            </div>
            <div>
              <div class="text-xs text-muted-foreground">
                Phone
              </div>
              <div class="text-sm">
                {{ selected?.phoneNumber || 'N/A' }}
              </div>
            </div>
          </div>

          <div v-if="isAdmin" class="rounded-md border p-3">
            <div class="flex items-center justify-between">
              <div>
                <div class="text-xs text-muted-foreground">
                  System Login
                </div>
                <Badge :variant="loginStatusVariant(selected?.loginStatus ?? null)" class="mt-1">
                  {{ loginStatusLabel(selected?.loginStatus ?? null) }}
                </Badge>
              </div>
              <Button
                v-if="selected?.loginStatus === 'pending'"
                size="sm"
                variant="outline"
                :disabled="isResending"
                @click="selected && resendInvite(selected)"
              >
                Resend Invite
              </Button>
            </div>
          </div>

          <div class="pt-2 flex items-center gap-2">
            <Button variant="secondary" @click="open = false">
              Close
            </Button>
            <AlertDialog v-if="isAdmin">
              <AlertDialogTrigger as-child>
                <Button variant="destructive" :disabled="isDeleting">
                  Delete Staff Member
                </Button>
              </AlertDialogTrigger>
              <AlertDialogContent>
                <AlertDialogHeader>
                  <AlertDialogTitle>Delete {{ selected?.firstName }} {{ selected?.lastName }}?</AlertDialogTitle>
                  <AlertDialogDescription>
                    This removes their staff record. If they have system login access, it will be revoked immediately. This cannot be undone.
                  </AlertDialogDescription>
                </AlertDialogHeader>
                <AlertDialogFooter>
                  <AlertDialogCancel>Cancel</AlertDialogCancel>
                  <AlertDialogAction @click="selected && deleteStaff(selected)">
                    Delete
                  </AlertDialogAction>
                </AlertDialogFooter>
              </AlertDialogContent>
            </AlertDialog>
          </div>
        </div>
      </SheetContent>
    </Sheet>
  </div>
</template>
