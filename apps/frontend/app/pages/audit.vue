<script setup lang="ts">
definePageMeta({
  middleware: [() => {
    const auth = useAuthStore()
    if (!auth.isAdmin) return navigateTo('/')
  }],
})

interface AuditEntry {
  id: string
  actor_user_id: string | null
  actor_name: string | null
  entity: string
  entity_id: string
  action: 'create' | 'update' | 'delete'
  before: unknown
  after: unknown
  created_at: string
}

const entityOptions = [
  { value: 'all', label: 'All entities' },
  { value: 'account', label: 'Account' },
  { value: 'contact', label: 'Contact' },
  { value: 'lead', label: 'Lead' },
  { value: 'case', label: 'Case' },
]

const api = useApi()
const entries = ref<AuditEntry[]>([])
const error = ref('')
const loading = ref(true)
const entityFilter = ref('all')

async function load() {
  error.value = ''
  loading.value = true
  try {
    const query = entityFilter.value !== 'all' ? `?entity=${entityFilter.value}` : ''
    entries.value = await api<AuditEntry[]>(`/audit${query}`) ?? []
  } catch {
    error.value = 'Could not load the audit log.'
  } finally {
    loading.value = false
  }
}

watch(entityFilter, load)

const diffOpen = ref(false)
const diffFor = ref<AuditEntry | null>(null)
function openDiff(entry: AuditEntry) {
  diffFor.value = entry
  diffOpen.value = true
}

function formatTime(iso: string) {
  return new Date(iso).toLocaleString()
}

const actionVariant: Record<string, 'default' | 'secondary' | 'destructive'> = {
  create: 'default',
  update: 'secondary',
  delete: 'destructive',
}

onMounted(load)
</script>

<template>
  <div class="flex flex-col gap-6">
    <div class="flex items-center justify-between">
      <div>
        <h1 class="text-2xl font-semibold tracking-tight">Audit Log</h1>
        <p class="text-muted-foreground">
          Every create/update/delete on accounts, contacts, leads, and cases — who, when, and what changed.
        </p>
      </div>
      <Select v-model="entityFilter">
        <SelectTrigger class="w-48"><SelectValue /></SelectTrigger>
        <SelectContent>
          <SelectItem v-for="e in entityOptions" :key="e.value" :value="e.value">{{ e.label }}</SelectItem>
        </SelectContent>
      </Select>
    </div>

    <p v-if="error" class="text-sm text-destructive">{{ error }}</p>

    <Card>
      <CardContent>
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead>When</TableHead>
              <TableHead>Actor</TableHead>
              <TableHead>Entity</TableHead>
              <TableHead>Record</TableHead>
              <TableHead>Action</TableHead>
              <TableHead class="text-right">Details</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            <TableRow v-if="!loading && entries.length === 0">
              <TableCell colspan="6" class="text-center text-muted-foreground py-8">
                No audit entries yet.
              </TableCell>
            </TableRow>
            <TableRow v-for="entry in entries" :key="entry.id">
              <TableCell class="text-muted-foreground">{{ formatTime(entry.created_at) }}</TableCell>
              <TableCell>{{ entry.actor_name ?? 'System' }}</TableCell>
              <TableCell class="capitalize">{{ entry.entity }}</TableCell>
              <TableCell class="font-mono text-xs text-muted-foreground">{{ entry.entity_id.slice(0, 8) }}…</TableCell>
              <TableCell>
                <Badge :variant="actionVariant[entry.action] ?? 'default'" class="capitalize">{{ entry.action }}</Badge>
              </TableCell>
              <TableCell class="text-right">
                <Button variant="ghost" size="sm" @click="openDiff(entry)">View</Button>
              </TableCell>
            </TableRow>
          </TableBody>
        </Table>
      </CardContent>
    </Card>

    <Dialog v-model:open="diffOpen">
      <DialogContent class="sm:max-w-2xl">
        <DialogHeader>
          <DialogTitle>{{ diffFor?.entity }} — {{ diffFor?.action }}</DialogTitle>
          <DialogDescription>
            {{ diffFor?.actor_name ?? 'System' }} · {{ diffFor ? formatTime(diffFor.created_at) : '' }}
          </DialogDescription>
        </DialogHeader>
        <div class="grid grid-cols-2 gap-4 text-sm">
          <div>
            <p class="mb-1 font-medium text-muted-foreground">Before</p>
            <pre class="max-h-80 overflow-auto rounded-md border border-border bg-muted p-2 text-xs">{{ diffFor?.before ? JSON.stringify(diffFor.before, null, 2) : '—' }}</pre>
          </div>
          <div>
            <p class="mb-1 font-medium text-muted-foreground">After</p>
            <pre class="max-h-80 overflow-auto rounded-md border border-border bg-muted p-2 text-xs">{{ diffFor?.after ? JSON.stringify(diffFor.after, null, 2) : '—' }}</pre>
          </div>
        </div>
      </DialogContent>
    </Dialog>
  </div>
</template>
