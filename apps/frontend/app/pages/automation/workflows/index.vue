<script setup lang="ts">
definePageMeta({
  middleware: [() => {
    const auth = useAuthStore()
    if (!auth.isAdmin) return navigateTo('/')
  }],
})

interface Condition {
  field: string
  operator: 'equals' | 'not_equals' | 'changed'
  value: string
}
interface Action {
  type: 'update_field' | 'assign_owner' | 'notify' | 'webhook'
  field?: string
  value?: string
  user_id?: string
  title?: string
  body?: string
  url?: string
}
interface Workflow {
  id: string
  name: string
  entity: string
  trigger_event: string
  conditions: Condition[]
  actions: Action[]
  is_active: boolean
  created_at: string
}
interface WorkflowRun {
  id: string
  record_id: string
  status: string
  detail: string
  ran_at: string
}

const entityOptions = [
  { value: 'account', label: 'Account' },
  { value: 'contact', label: 'Contact' },
  { value: 'lead', label: 'Lead' },
  { value: 'case', label: 'Case' },
]
const triggerOptions = [
  { value: 'created', label: 'Record created' },
  { value: 'updated', label: 'Record updated' },
  { value: 'sla_breach', label: 'SLA breached (cases only)' },
]
const operatorOptions = [
  { value: 'equals', label: 'equals' },
  { value: 'not_equals', label: 'does not equal' },
  { value: 'changed', label: 'changed' },
]
const actionTypeOptions = [
  { value: 'update_field', label: 'Update a field' },
  { value: 'assign_owner', label: 'Reassign owner' },
  { value: 'notify', label: 'Send in-app notification' },
  { value: 'webhook', label: 'Call a webhook' },
]
// Fields available per entity for conditions (mirrors the newFields the engine builds).
const conditionFields: Record<string, string[]> = {
  account: ['name', 'owner_user_id'],
  contact: ['first_name', 'last_name', 'email', 'phone', 'owner_user_id'],
  lead: ['topic', 'status', 'owner_user_id'],
  case: ['subject', 'priority', 'status', 'owner_user_id', 'breach_type'],
}
// Fields update_field is allowed to write, per entity (mirrors fieldSetters in the engine).
const updatableFields: Record<string, string[]> = {
  account: ['name'],
  contact: ['first_name', 'last_name', 'email', 'phone'],
  lead: ['topic', 'status'],
  case: ['subject', 'priority', 'status'],
}

const api = useApi()
const workflows = ref<Workflow[]>([])
const error = ref('')
const loading = ref(true)

async function load() {
  error.value = ''
  try {
    workflows.value = await api<Workflow[]>('/automation/workflows') ?? []
  } catch {
    error.value = 'Could not load workflows.'
  } finally {
    loading.value = false
  }
}

function emptyForm(): Workflow {
  return {
    id: '', name: '', entity: 'lead', trigger_event: 'created',
    conditions: [], actions: [], is_active: true, created_at: '',
  }
}

const editorOpen = ref(false)
const editing = ref<Workflow>(emptyForm())
const isNew = computed(() => !editing.value.id)
const saving = ref(false)
const saveError = ref('')

function openCreate() {
  editing.value = emptyForm()
  saveError.value = ''
  editorOpen.value = true
}

function openEdit(wf: Workflow) {
  editing.value = JSON.parse(JSON.stringify(wf))
  saveError.value = ''
  editorOpen.value = true
}

function addCondition() {
  editing.value.conditions.push({ field: conditionFields[editing.value.entity]?.[0] ?? '', operator: 'equals', value: '' })
}
function removeCondition(index: number) {
  editing.value.conditions.splice(index, 1)
}
function addAction() {
  editing.value.actions.push({ type: 'notify', user_id: 'owner', title: '', body: '' })
}
function removeAction(index: number) {
  editing.value.actions.splice(index, 1)
}

async function save() {
  saving.value = true
  saveError.value = ''
  try {
    const body = {
      name: editing.value.name,
      entity: editing.value.entity,
      trigger_event: editing.value.trigger_event,
      conditions: editing.value.conditions,
      actions: editing.value.actions,
      is_active: editing.value.is_active,
    }
    if (isNew.value) {
      await api('/automation/workflows', { method: 'POST', body })
    } else {
      await api(`/automation/workflows/${editing.value.id}`, { method: 'PUT', body })
    }
    editorOpen.value = false
    await load()
  } catch {
    saveError.value = 'Could not save workflow. Check that every condition/action has the required fields filled in.'
  } finally {
    saving.value = false
  }
}

async function toggleActive(wf: Workflow) {
  try {
    await api(`/automation/workflows/${wf.id}`, {
      method: 'PUT',
      body: { name: wf.name, entity: wf.entity, trigger_event: wf.trigger_event, conditions: wf.conditions, actions: wf.actions, is_active: !wf.is_active },
    })
    await load()
  } catch {
    error.value = `Could not update "${wf.name}".`
  }
}

async function deleteWorkflow(wf: Workflow) {
  try {
    await api(`/automation/workflows/${wf.id}`, { method: 'DELETE' })
    await load()
  } catch {
    error.value = `Could not delete "${wf.name}".`
  }
}

const runsOpen = ref(false)
const runsFor = ref<Workflow | null>(null)
const runs = ref<WorkflowRun[]>([])
const runsError = ref('')

async function openRuns(wf: Workflow) {
  runsFor.value = wf
  runsOpen.value = true
  runsError.value = ''
  runs.value = []
  try {
    runs.value = await api<WorkflowRun[]>(`/automation/workflows/${wf.id}/runs`) ?? []
  } catch {
    runsError.value = 'Could not load run history.'
  }
}

function formatTime(iso: string) {
  return new Date(iso).toLocaleString()
}

onMounted(load)
</script>

<template>
  <div class="flex flex-col gap-6">
    <div class="flex items-center justify-between">
      <div>
        <h1 class="text-2xl font-semibold tracking-tight">Workflows</h1>
        <p class="text-muted-foreground">
          Automate what happens when a record is created or updated, or an SLA is breached.
        </p>
      </div>
      <Button @click="openCreate">
        <Icon name="lucide:plus" class="mr-1.5 size-4" />
        New workflow
      </Button>
    </div>

    <p v-if="error" class="text-sm text-destructive">{{ error }}</p>

    <Card>
      <CardContent>
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead>Name</TableHead>
              <TableHead>Entity</TableHead>
              <TableHead>Trigger</TableHead>
              <TableHead>Conditions</TableHead>
              <TableHead>Actions</TableHead>
              <TableHead>Status</TableHead>
              <TableHead class="text-right">Manage</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            <TableRow v-if="!loading && workflows.length === 0">
              <TableCell colspan="7" class="text-center text-muted-foreground py-8">
                No workflows yet. Create one to react automatically to record changes or SLA breaches.
              </TableCell>
            </TableRow>
            <TableRow v-for="wf in workflows" :key="wf.id">
              <TableCell class="font-medium">{{ wf.name }}</TableCell>
              <TableCell class="capitalize">{{ wf.entity }}</TableCell>
              <TableCell>{{ triggerOptions.find((t) => t.value === wf.trigger_event)?.label ?? wf.trigger_event }}</TableCell>
              <TableCell class="text-muted-foreground">{{ wf.conditions.length }}</TableCell>
              <TableCell class="text-muted-foreground">{{ wf.actions.length }}</TableCell>
              <TableCell>
                <Badge :variant="wf.is_active ? 'default' : 'outline'" class="cursor-pointer" @click="toggleActive(wf)">
                  {{ wf.is_active ? 'Active' : 'Paused' }}
                </Badge>
              </TableCell>
              <TableCell class="text-right space-x-1">
                <Button variant="ghost" size="sm" @click="openRuns(wf)">Runs</Button>
                <Button variant="ghost" size="sm" @click="openEdit(wf)">Edit</Button>
                <Button variant="ghost" size="sm" class="text-destructive" @click="deleteWorkflow(wf)">Delete</Button>
              </TableCell>
            </TableRow>
          </TableBody>
        </Table>
      </CardContent>
    </Card>

    <Dialog v-model:open="editorOpen">
      <DialogScrollContent class="sm:max-w-2xl">
        <DialogHeader>
          <DialogTitle>{{ isNew ? 'Create a workflow' : 'Edit workflow' }}</DialogTitle>
          <DialogDescription>Trigger → conditions → actions. All conditions must match for the actions to run.</DialogDescription>
        </DialogHeader>

        <form class="space-y-6" @submit.prevent="save">
          <div class="grid grid-cols-1 sm:grid-cols-3 gap-4">
            <div class="space-y-1.5 sm:col-span-1">
              <Label for="wf-name">Name</Label>
              <Input id="wf-name" v-model="editing.name" required />
            </div>
            <div class="space-y-1.5">
              <Label>Entity</Label>
              <Select v-model="editing.entity">
                <SelectTrigger><SelectValue /></SelectTrigger>
                <SelectContent>
                  <SelectItem v-for="e in entityOptions" :key="e.value" :value="e.value">{{ e.label }}</SelectItem>
                </SelectContent>
              </Select>
            </div>
            <div class="space-y-1.5">
              <Label>Trigger</Label>
              <Select v-model="editing.trigger_event">
                <SelectTrigger><SelectValue /></SelectTrigger>
                <SelectContent>
                  <SelectItem v-for="t in triggerOptions" :key="t.value" :value="t.value">{{ t.label }}</SelectItem>
                </SelectContent>
              </Select>
            </div>
          </div>

          <div class="space-y-2">
            <div class="flex items-center justify-between">
              <Label>Conditions</Label>
              <Button type="button" variant="outline" size="sm" @click="addCondition">
                <Icon name="lucide:plus" class="mr-1 size-3.5" /> Add condition
              </Button>
            </div>
            <p v-if="editing.conditions.length === 0" class="text-sm text-muted-foreground">
              No conditions — the actions always run on this trigger.
            </p>
            <div v-for="(cond, i) in editing.conditions" :key="i" class="flex items-center gap-2">
              <Select v-model="cond.field">
                <SelectTrigger class="w-40"><SelectValue /></SelectTrigger>
                <SelectContent>
                  <SelectItem v-for="f in conditionFields[editing.entity]" :key="f" :value="f">{{ f }}</SelectItem>
                </SelectContent>
              </Select>
              <Select v-model="cond.operator">
                <SelectTrigger class="w-40"><SelectValue /></SelectTrigger>
                <SelectContent>
                  <SelectItem v-for="o in operatorOptions" :key="o.value" :value="o.value">{{ o.label }}</SelectItem>
                </SelectContent>
              </Select>
              <Input v-if="cond.operator !== 'changed'" v-model="cond.value" placeholder="value" class="flex-1" />
              <div v-else class="flex-1" />
              <Button type="button" variant="ghost" size="icon" @click="removeCondition(i)">
                <Icon name="lucide:x" class="size-4" />
              </Button>
            </div>
          </div>

          <div class="space-y-2">
            <div class="flex items-center justify-between">
              <Label>Actions</Label>
              <Button type="button" variant="outline" size="sm" @click="addAction">
                <Icon name="lucide:plus" class="mr-1 size-3.5" /> Add action
              </Button>
            </div>
            <p v-if="editing.actions.length === 0" class="text-sm text-destructive">
              Add at least one action or this workflow will do nothing.
            </p>
            <div v-for="(action, i) in editing.actions" :key="i" class="rounded-md border border-border p-3 space-y-2">
              <div class="flex items-center gap-2">
                <Select v-model="action.type">
                  <SelectTrigger class="w-56"><SelectValue /></SelectTrigger>
                  <SelectContent>
                    <SelectItem v-for="t in actionTypeOptions" :key="t.value" :value="t.value">{{ t.label }}</SelectItem>
                  </SelectContent>
                </Select>
                <div class="flex-1" />
                <Button type="button" variant="ghost" size="icon" @click="removeAction(i)">
                  <Icon name="lucide:x" class="size-4" />
                </Button>
              </div>

              <div v-if="action.type === 'update_field'" class="grid grid-cols-2 gap-2">
                <Select v-model="action.field">
                  <SelectTrigger><SelectValue placeholder="Field" /></SelectTrigger>
                  <SelectContent>
                    <SelectItem v-for="f in updatableFields[editing.entity]" :key="f" :value="f">{{ f }}</SelectItem>
                  </SelectContent>
                </Select>
                <Input v-model="action.value" placeholder="New value" />
              </div>

              <div v-else-if="action.type === 'assign_owner'" class="space-y-1">
                <Input v-model="action.user_id" placeholder="User ID to assign as owner" />
              </div>

              <div v-else-if="action.type === 'notify'" class="space-y-2">
                <Input v-model="action.user_id" placeholder="User ID, or &quot;owner&quot; for the record's owner" />
                <Input v-model="action.title" placeholder="Notification title" />
                <Textarea v-model="action.body" placeholder="Notification body" rows="2" />
              </div>

              <div v-else-if="action.type === 'webhook'" class="space-y-1">
                <Input v-model="action.url" placeholder="https://example.com/webhook" />
              </div>
            </div>
          </div>

          <p v-if="saveError" class="text-sm text-destructive">{{ saveError }}</p>
          <DialogFooter>
            <Button type="submit" :disabled="saving">{{ saving ? 'Saving…' : 'Save workflow' }}</Button>
          </DialogFooter>
        </form>
      </DialogScrollContent>
    </Dialog>

    <Dialog v-model:open="runsOpen">
      <DialogContent class="sm:max-w-xl">
        <DialogHeader>
          <DialogTitle>Run history — {{ runsFor?.name }}</DialogTitle>
          <DialogDescription>The last 50 times this workflow's conditions matched and its actions ran.</DialogDescription>
        </DialogHeader>
        <p v-if="runsError" class="text-sm text-destructive">{{ runsError }}</p>
        <p v-else-if="runs.length === 0" class="text-sm text-muted-foreground">No runs yet.</p>
        <div v-else class="max-h-96 overflow-y-auto space-y-2">
          <div v-for="run in runs" :key="run.id" class="rounded-md border border-border p-2 text-sm">
            <div class="flex items-center justify-between">
              <Badge :variant="run.status === 'success' ? 'default' : 'destructive'">{{ run.status }}</Badge>
              <span class="text-xs text-muted-foreground">{{ formatTime(run.ran_at) }}</span>
            </div>
            <p class="mt-1 text-muted-foreground">{{ run.detail }}</p>
          </div>
        </div>
      </DialogContent>
    </Dialog>
  </div>
</template>
