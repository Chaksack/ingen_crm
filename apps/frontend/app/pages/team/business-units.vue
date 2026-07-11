<script setup lang="ts">
definePageMeta({
  middleware: [() => {
    const auth = useAuthStore()
    if (!auth.isAdmin) return navigateTo('/')
  }],
})

interface BU {
  id: string
  name: string
  parent_id: string | null
}
interface BURow extends BU {
  depth: number
}

const api = useApi()
const units = ref<BU[]>([])
const error = ref('')

const dialogOpen = ref(false)
const editingId = ref<string | null>(null)
const formName = ref('')
const formParentId = ref<string>('none')
const submitting = ref(false)
const formError = ref('')

const rows = computed<BURow[]>(() => {
  const byParent = new Map<string | null, BU[]>()
  for (const u of units.value) {
    const key = u.parent_id
    if (!byParent.has(key)) byParent.set(key, [])
    byParent.get(key)!.push(u)
  }
  const out: BURow[] = []
  function walk(parentId: string | null, depth: number) {
    for (const u of byParent.get(parentId) ?? []) {
      out.push({ ...u, depth })
      walk(u.id, depth + 1)
    }
  }
  walk(null, 0)
  return out
})

async function load() {
  error.value = ''
  try {
    units.value = await api<BU[]>('/team/business-units') ?? []
  } catch {
    error.value = 'Could not load business units.'
  }
}

function openCreate() {
  editingId.value = null
  formName.value = ''
  formParentId.value = 'none'
  formError.value = ''
  dialogOpen.value = true
}

function openEdit(bu: BU) {
  editingId.value = bu.id
  formName.value = bu.name
  formParentId.value = bu.parent_id ?? 'none'
  formError.value = ''
  dialogOpen.value = true
}

async function submit() {
  if (!formName.value.trim()) return
  submitting.value = true
  formError.value = ''
  const body = { name: formName.value, parent_id: formParentId.value === 'none' ? null : formParentId.value }
  try {
    if (editingId.value) {
      await api(`/team/business-units/${editingId.value}`, { method: 'PATCH', body })
    } else {
      await api('/team/business-units', { method: 'POST', body })
    }
    dialogOpen.value = false
    await load()
  } catch {
    formError.value = 'Could not save. Check that this would not create a cycle in the hierarchy.'
  } finally {
    submitting.value = false
  }
}

async function remove(bu: BU) {
  error.value = ''
  try {
    await api(`/team/business-units/${bu.id}`, { method: 'DELETE' })
    await load()
  } catch {
    error.value = `Could not delete "${bu.name}" — move its members and any child business units first.`
  }
}

onMounted(load)
</script>

<template>
  <div class="flex flex-col gap-6">
    <div class="flex items-center justify-between">
      <div>
        <h1 class="text-2xl font-semibold tracking-tight">Business Units</h1>
        <p class="text-muted-foreground">The hierarchy privilege depth (BU / BU + sub-units) is scoped against.</p>
      </div>
      <Dialog v-model:open="dialogOpen">
        <DialogTrigger as-child>
          <Button @click="openCreate">
            <Icon name="lucide:plus" class="mr-1.5 size-4" />
            New business unit
          </Button>
        </DialogTrigger>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>{{ editingId ? 'Rename business unit' : 'Create business unit' }}</DialogTitle>
          </DialogHeader>
          <form class="space-y-4" @submit.prevent="submit">
            <div class="space-y-1.5">
              <Label for="bu-name">Name</Label>
              <Input id="bu-name" v-model="formName" required />
            </div>
            <div class="space-y-1.5">
              <Label for="bu-parent">Parent</Label>
              <Select v-model="formParentId">
                <SelectTrigger id="bu-parent"><SelectValue /></SelectTrigger>
                <SelectContent>
                  <SelectItem value="none">None (root)</SelectItem>
                  <SelectItem
                    v-for="u in units.filter((u) => u.id !== editingId)"
                    :key="u.id"
                    :value="u.id"
                  >
                    {{ u.name }}
                  </SelectItem>
                </SelectContent>
              </Select>
            </div>
            <p v-if="formError" class="text-sm text-destructive">{{ formError }}</p>
            <Button type="submit" class="w-full" :disabled="submitting">
              {{ submitting ? 'Saving…' : 'Save' }}
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
              <TableHead class="text-right">Actions</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            <TableRow v-for="bu in rows" :key="bu.id">
              <TableCell>
                <span :style="{ paddingLeft: `${bu.depth * 1.5}rem` }" class="inline-flex items-center gap-1.5">
                  <Icon v-if="bu.depth > 0" name="lucide:corner-down-right" class="size-3.5 text-muted-foreground" />
                  {{ bu.name }}
                  <Badge v-if="!bu.parent_id" variant="outline" class="ml-1">Root</Badge>
                </span>
              </TableCell>
              <TableCell class="text-right">
                <Button variant="ghost" size="sm" @click="openEdit(bu)">Edit</Button>
                <Button v-if="bu.parent_id" variant="ghost" size="sm" class="text-destructive" @click="remove(bu)">
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
