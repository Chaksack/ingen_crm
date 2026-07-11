<script setup lang="ts">
definePageMeta({
  middleware: [() => {
    const auth = useAuthStore()
    if (!auth.isAdmin) return navigateTo('/')
  }],
})

interface RoleRow {
  id: string
  name: string
  member_count: number
  builtin: boolean
}

const api = useApi()
const roles = ref<RoleRow[]>([])
const error = ref('')

const createOpen = ref(false)
const newRoleName = ref('')
const creating = ref(false)
const createError = ref('')

async function load() {
  error.value = ''
  try {
    roles.value = await api<RoleRow[]>('/team/roles') ?? []
  } catch {
    error.value = 'Could not load roles.'
  }
}

async function createRole() {
  if (!newRoleName.value.trim()) return
  creating.value = true
  createError.value = ''
  try {
    await api('/team/roles', { method: 'POST', body: { name: newRoleName.value } })
    newRoleName.value = ''
    createOpen.value = false
    await load()
  } catch {
    createError.value = 'Could not create role. A role with that name may already exist.'
  } finally {
    creating.value = false
  }
}

async function deleteRole(role: RoleRow) {
  try {
    await api(`/team/roles/${role.id}`, { method: 'DELETE' })
    await load()
  } catch {
    error.value = `Could not delete "${role.name}". Reassign its members to another role first.`
  }
}

onMounted(load)
</script>

<template>
  <div class="flex flex-col gap-6">
    <div class="flex items-center justify-between">
      <div>
        <h1 class="text-2xl font-semibold tracking-tight">Roles</h1>
        <p class="text-muted-foreground">Security roles and their per-entity privilege depth.</p>
      </div>
      <Dialog v-model:open="createOpen">
        <DialogTrigger as-child>
          <Button>
            <Icon name="lucide:plus" class="mr-1.5 size-4" />
            New role
          </Button>
        </DialogTrigger>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Create a role</DialogTitle>
            <DialogDescription>Starts with no access to anything — grant privileges from the role's page.</DialogDescription>
          </DialogHeader>
          <form class="space-y-4" @submit.prevent="createRole">
            <div class="space-y-1.5">
              <Label for="role-name">Role name</Label>
              <Input id="role-name" v-model="newRoleName" required />
            </div>
            <p v-if="createError" class="text-sm text-destructive">{{ createError }}</p>
            <Button type="submit" class="w-full" :disabled="creating">
              {{ creating ? 'Creating…' : 'Create role' }}
            </Button>
          </form>
        </DialogContent>
      </Dialog>
    </div>

    <p v-if="error" class="text-sm text-destructive">{{ error }}</p>

    <Card>
      <CardContent>
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead>Name</TableHead>
              <TableHead>Members</TableHead>
              <TableHead class="text-right">Actions</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            <TableRow v-for="role in roles" :key="role.id">
              <TableCell>
                <NuxtLink :to="`/team/roles/${role.id}`" class="font-medium hover:underline">
                  {{ role.name }}
                </NuxtLink>
                <Badge v-if="role.builtin" variant="outline" class="ml-2">Built-in</Badge>
              </TableCell>
              <TableCell class="text-muted-foreground">{{ role.member_count }}</TableCell>
              <TableCell class="text-right">
                <Button as-child variant="ghost" size="sm">
                  <NuxtLink :to="`/team/roles/${role.id}`">Edit privileges</NuxtLink>
                </Button>
                <Button
                  v-if="!role.builtin"
                  variant="ghost"
                  size="sm"
                  class="text-destructive"
                  @click="deleteRole(role)"
                >
                  Delete
                </Button>
              </TableCell>
            </TableRow>
          </TableBody>
        </Table>
      </CardContent>
    </Card>
  </div>
</template>
