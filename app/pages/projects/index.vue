<script setup lang="ts">
import { useRouter } from 'vue-router'
import { useProjects } from '~/composables/useProjects'

const { projects, members, createProject, removeProject } = useProjects()
const router = useRouter()

const form = reactive({
  name: '',
  key: '',
  description: '',
  leadId: '',
  memberIds: [] as string[],
  customerName: '',
})

function resetForm() {
  form.name = ''
  form.key = ''
  form.description = ''
  form.leadId = members.value[0]?.id || ''
  form.memberIds = form.leadId ? [form.leadId] : []
  form.customerName = ''
}

const open = ref(false)

function handleCreate() {
  if (!form.name.trim())
    return
  const p = createProject({
    name: form.name,
    key: form.key,
    description: form.description,
    leadId: form.leadId || members.value[0]?.id,
    memberIds: form.memberIds,
    customerName: form.customerName,
  })
  open.value = false
  resetForm()
  router.push(`/projects/${p.id}`)
}

onMounted(() => {
  if (!form.leadId)
    form.leadId = members.value[0]?.id || ''
  if (!form.memberIds.length && form.leadId)
    form.memberIds = [form.leadId]
})
</script>

<template>
  <div class="w-full flex flex-col gap-4">
    <div class="flex flex-wrap items-center justify-between gap-2">
      <h2 class="text-2xl font-bold tracking-tight">Projects</h2>
      <div class="flex items-center space-x-2">
        <Sheet v-model:open="open">
          <SheetTrigger as-child>
            <Button>
              <Icon name="i-lucide-plus" />
              New Project
            </Button>
          </SheetTrigger>
          <SheetContent class="w-full sm:w-3/4 lg:w-1/2 rounded-l-lg sm:max-w-none overflow-y-auto p-6">
            <SheetHeader>
              <SheetTitle>Create New Project</SheetTitle>
              <SheetDescription>
                Provide details to create a new project.
              </SheetDescription>
            </SheetHeader>
            <div class="space-y-4 py-4">
              <div class="space-y-2">
                <Label for="pname">Name</Label>
                <Input id="pname" v-model="form.name" placeholder="Website Revamp" />
              </div>
              <div class="space-y-2">
                <Label for="pkey">Key</Label>
                <Input id="pkey" v-model="form.key" placeholder="WEB" />
                <p class="text-xs text-muted-foreground">Short code used for task keys.</p>
              </div>
              <div class="space-y-2">
                <Label for="pcustomer">Customer / Company</Label>
                <Input id="pcustomer" v-model="form.customerName" placeholder="Acme Corp" />
              </div>
              <div class="space-y-2">
                <Label for="pdesc">Description</Label>
                <Textarea id="pdesc" v-model="form.description" rows="4" placeholder="Brief description..." />
              </div>
              <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
                <div class="space-y-2">
                  <Label for="lead">Project Lead</Label>
                  <Select v-model="form.leadId">
                    <SelectTrigger id="lead">
                      <SelectValue placeholder="Select lead" />
                    </SelectTrigger>
                    <SelectContent>
                      <SelectItem v-for="m in members" :key="m.id" :value="m.id">{{ m.name }}</SelectItem>
                    </SelectContent>
                  </Select>
                </div>
                <div class="space-y-2">
                  <Label>Members</Label>
                  <div class="flex flex-wrap gap-2">
                    <label v-for="m in members" :key="m.id" class="flex items-center gap-2 text-sm">
                      <Checkbox :checked="form.memberIds.includes(m.id)" @update:checked="(v: boolean) => { const set = new Set(form.memberIds); v ? set.add(m.id) : set.delete(m.id); form.memberIds = [...set] }" />
                      <span>{{ m.name }}</span>
                    </label>
                  </div>
                </div>
              </div>
              <div class="flex items-center gap-2">
                <Button :disabled="!form.name.trim()" @click="handleCreate">Create Project</Button>
                <Button variant="secondary" @click="open = false">Cancel</Button>
              </div>
            </div>
          </SheetContent>
        </Sheet>
      </div>
    </div>

    <main>
      <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
        <Card v-for="p in projects" :key="p.id" class="cursor-pointer hover:shadow" @click="router.push(`/projects/${p.id}`)">
          <CardHeader>
            <CardTitle class="flex items-center justify-between">
              <span>{{ p.name }}</span>
              <Badge variant="secondary">{{ p.key }}</Badge>
            </CardTitle>
            <div v-if="p.customerName" class="text-sm text-muted-foreground">Customer: {{ p.customerName }}</div>
            <CardDescription class="line-clamp-2">{{ p.description }}</CardDescription>
          </CardHeader>
          <CardFooter class="flex items-center justify-between">
            <div class="text-xs text-muted-foreground">Created {{ new Date(p.createdAt).toLocaleDateString() }}</div>
            <Button size="sm" variant="destructive" @click.stop="removeProject(p.id)">
              <Icon name="i-lucide-trash-2" />
              Delete
            </Button>
          </CardFooter>
        </Card>
        <div v-if="projects.length === 0" class="col-span-full text-center text-muted-foreground py-12">
          No projects yet. Create your first project.
        </div>
      </div>
    </main>
  </div>
</template>
