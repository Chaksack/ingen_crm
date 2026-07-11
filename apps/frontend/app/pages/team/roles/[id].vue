<script setup lang="ts">
definePageMeta({
  middleware: [() => {
    const auth = useAuthStore()
    if (!auth.isAdmin) return navigateTo('/')
  }],
})

interface PrivilegeEntry {
  entity: string
  action: string
  depth: string
}
interface RoleRow {
  id: string
  name: string
  member_count: number
  builtin: boolean
}

const route = useRoute()
const roleId = route.params.id as string
const api = useApi()
const breadcrumbOverride = useBreadcrumbOverride()

const entities = [
  { key: 'account', label: 'Accounts' },
  { key: 'contact', label: 'Contacts' },
  { key: 'lead', label: 'Leads' },
  { key: 'case', label: 'Cases' },
]
const actions = [
  { key: 'create', label: 'Create' },
  { key: 'read', label: 'Read' },
  { key: 'write', label: 'Write' },
  { key: 'delete', label: 'Delete' },
  { key: 'assign', label: 'Assign' },
]
const depthOptions = [
  { value: 'none', label: 'No access' },
  { value: 'user', label: 'Own records' },
  { value: 'bu', label: 'My business unit' },
  { value: 'bu_children', label: 'My BU + sub-units' },
  { value: 'org', label: 'Entire organization' },
]

const role = ref<RoleRow | null>(null)
const grid = reactive<Record<string, string>>({})
const saving = ref(false)
const error = ref('')
const saved = ref(false)

function cellKey(entity: string, action: string) {
  return `${entity}:${action}`
}

async function load() {
  const [roles, privileges] = await Promise.all([
    api<RoleRow[]>('/team/roles'),
    api<PrivilegeEntry[]>(`/team/roles/${roleId}/privileges`),
  ])
  role.value = (roles ?? []).find((r) => r.id === roleId) ?? null
  breadcrumbOverride.value = role.value?.name ?? null
  for (const p of privileges ?? []) {
    grid[cellKey(p.entity, p.action)] = p.depth
  }
}

async function save() {
  saving.value = true
  saved.value = false
  error.value = ''
  try {
    const entries: PrivilegeEntry[] = []
    for (const e of entities) {
      for (const a of actions) {
        entries.push({ entity: e.key, action: a.key, depth: grid[cellKey(e.key, a.key)] ?? 'none' })
      }
    }
    await api(`/team/roles/${roleId}/privileges`, { method: 'PUT', body: entries })
    saved.value = true
  } catch {
    error.value = 'Could not save privileges.'
  } finally {
    saving.value = false
  }
}

onMounted(load)
</script>

<template>
  <div class="flex flex-col gap-6">
    <div>
      <h1 class="text-2xl font-semibold tracking-tight">
        {{ role?.name }}
        <Badge v-if="role?.builtin" variant="outline" class="ml-2 align-middle">Built-in</Badge>
      </h1>
      <p class="text-muted-foreground">
        For each entity and action, choose the deepest scope this role can act on.
      </p>
    </div>

    <Card>
      <CardContent>
        <div class="overflow-x-auto">
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>Entity</TableHead>
                <TableHead v-for="a in actions" :key="a.key">{{ a.label }}</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              <TableRow v-for="e in entities" :key="e.key">
                <TableCell class="font-medium">{{ e.label }}</TableCell>
                <TableCell v-for="a in actions" :key="a.key">
                  <Select v-model="grid[cellKey(e.key, a.key)]">
                    <SelectTrigger class="w-44"><SelectValue /></SelectTrigger>
                    <SelectContent>
                      <SelectItem v-for="opt in depthOptions" :key="opt.value" :value="opt.value">
                        {{ opt.label }}
                      </SelectItem>
                    </SelectContent>
                  </Select>
                </TableCell>
              </TableRow>
            </TableBody>
          </Table>
        </div>
      </CardContent>
    </Card>

    <div class="flex items-center gap-3">
      <Button :disabled="saving" @click="save">{{ saving ? 'Saving…' : 'Save privileges' }}</Button>
      <span v-if="saved" class="text-sm text-muted-foreground">Saved.</span>
      <span v-if="error" class="text-sm text-destructive">{{ error }}</span>
    </div>
  </div>
</template>
