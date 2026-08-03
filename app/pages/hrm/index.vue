<script setup lang="ts">
interface StaffItem {
  id: string
  firstName: string
  lastName: string
  status: 'active' | 'inactive'
  department?: string
  role: 'Staff' | 'Manager' | 'Admin'
}

const { data: staffData, pending: isLoading } = await useFetch<StaffItem[]>('/api/staff')
const staff = computed(() => staffData.value ?? [])

const activeCount = computed(() => staff.value.filter(s => s.status === 'active').length)
const inactiveCount = computed(() => staff.value.filter(s => s.status === 'inactive').length)

const departmentBreakdown = computed(() => {
  const counts = new Map<string, number>()
  for (const member of staff.value) {
    const key = member.department?.trim() || 'Unassigned'
    counts.set(key, (counts.get(key) ?? 0) + 1)
  }
  return Array.from(counts.entries())
    .map(([department, count]) => ({ department, count }))
    .sort((a, b) => b.count - a.count)
})
</script>

<template>
  <div class="w-full flex flex-col gap-4">
    <div class="flex flex-wrap items-center justify-between gap-2">
      <h2 class="text-2xl font-bold tracking-tight">
        Human Resource
      </h2>
      <NuxtLink to="/staff">
        <Button variant="outline">
          Manage Staff
          <Icon name="i-lucide-arrow-right" />
        </Button>
      </NuxtLink>
    </div>

    <div class="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
      <Card>
        <CardHeader class="pb-2">
          <CardDescription>Total Headcount</CardDescription>
          <CardTitle class="text-2xl">
            <Skeleton v-if="isLoading" class="h-8 w-16" />
            <span v-else>{{ staff.length }}</span>
          </CardTitle>
        </CardHeader>
      </Card>
      <Card>
        <CardHeader class="pb-2">
          <CardDescription>Active</CardDescription>
          <CardTitle class="text-2xl">
            <Skeleton v-if="isLoading" class="h-8 w-16" />
            <span v-else>{{ activeCount }}</span>
          </CardTitle>
        </CardHeader>
      </Card>
      <Card>
        <CardHeader class="pb-2">
          <CardDescription>Inactive</CardDescription>
          <CardTitle class="text-2xl">
            <Skeleton v-if="isLoading" class="h-8 w-16" />
            <span v-else>{{ inactiveCount }}</span>
          </CardTitle>
        </CardHeader>
      </Card>
    </div>

    <Card>
      <CardHeader>
        <CardTitle>Department Breakdown</CardTitle>
        <CardDescription>Headcount by department</CardDescription>
      </CardHeader>
      <CardContent>
        <div v-if="isLoading" class="space-y-2">
          <Skeleton class="h-6 w-full" />
          <Skeleton class="h-6 w-full" />
          <Skeleton class="h-6 w-full" />
        </div>
        <div v-else-if="departmentBreakdown.length === 0" class="text-sm text-muted-foreground">
          No staff records yet.
        </div>
        <div v-else class="space-y-3">
          <div v-for="item in departmentBreakdown" :key="item.department" class="flex items-center justify-between text-sm">
            <span>{{ item.department }}</span>
            <Badge variant="secondary">
              {{ item.count }}
            </Badge>
          </div>
        </div>
      </CardContent>
    </Card>
  </div>
</template>
