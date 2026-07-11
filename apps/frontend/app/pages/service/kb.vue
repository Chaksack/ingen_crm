<script setup lang="ts">
interface Article {
  id: string
  title: string
  status: string
  created_at: string
  updated_at: string
}

const api = useApi()
const auth = useAuthStore()

const articles = ref<Article[]>([])
const error = ref('')

const dialogOpen = ref(false)
const editingId = ref<string | null>(null)
const formTitle = ref('')
const formBody = ref('')
const submitting = ref(false)

async function load() {
  error.value = ''
  try {
    articles.value = await api<Article[]>('/service/kb') ?? []
  } catch {
    error.value = 'Could not load knowledge base articles.'
  }
}

function openCreate() {
  editingId.value = null
  formTitle.value = ''
  formBody.value = ''
  dialogOpen.value = true
}

async function openEdit(article: Article) {
  editingId.value = article.id
  const full = await api<{ title: string, body: string }>(`/service/kb/${article.id}`)
  formTitle.value = full.title
  formBody.value = full.body
  dialogOpen.value = true
}

async function submit() {
  if (!formTitle.value.trim()) return
  submitting.value = true
  try {
    if (editingId.value) {
      await api(`/service/kb/${editingId.value}`, { method: 'PUT', body: { title: formTitle.value, body: formBody.value } })
    } else {
      await api('/service/kb', { method: 'POST', body: { title: formTitle.value, body: formBody.value } })
    }
    dialogOpen.value = false
    await load()
  } finally {
    submitting.value = false
  }
}

async function togglePublish(article: Article) {
  const action = article.status === 'published' ? 'unpublish' : 'publish'
  await api(`/service/kb/${article.id}/${action}`, { method: 'POST' })
  await load()
}

async function remove(article: Article) {
  await api(`/service/kb/${article.id}`, { method: 'DELETE' })
  await load()
}

onMounted(load)
</script>

<template>
  <div class="flex flex-col gap-6">
    <div class="flex items-center justify-between">
      <div>
        <h1 class="text-2xl font-semibold tracking-tight">Knowledge Base</h1>
        <p class="text-muted-foreground">Articles agents can share, and (once published) that get suggested on cases.</p>
      </div>
      <Dialog v-model:open="dialogOpen">
        <DialogTrigger as-child>
          <Button @click="openCreate">
            <Icon name="lucide:plus" class="mr-1.5 size-4" />
            New article
          </Button>
        </DialogTrigger>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>{{ editingId ? 'Edit article' : 'New article' }}</DialogTitle>
          </DialogHeader>
          <form class="space-y-4" @submit.prevent="submit">
            <div class="space-y-1.5">
              <Label for="kb-title">Title</Label>
              <Input id="kb-title" v-model="formTitle" required />
            </div>
            <div class="space-y-1.5">
              <Label for="kb-body">Body</Label>
              <Textarea id="kb-body" v-model="formBody" rows="6" />
            </div>
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
              <TableHead>Title</TableHead>
              <TableHead>Status</TableHead>
              <TableHead>Updated</TableHead>
              <TableHead class="text-right">Actions</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            <TableRow v-for="a in articles" :key="a.id">
              <TableCell>
                <button class="text-left font-medium hover:underline" @click="openEdit(a)">{{ a.title }}</button>
              </TableCell>
              <TableCell>
                <Badge :variant="a.status === 'published' ? 'secondary' : 'outline'">{{ a.status }}</Badge>
              </TableCell>
              <TableCell class="text-muted-foreground">{{ new Date(a.updated_at).toLocaleString() }}</TableCell>
              <TableCell class="text-right">
                <template v-if="auth.isAdmin">
                  <Button variant="ghost" size="sm" @click="togglePublish(a)">
                    {{ a.status === 'published' ? 'Unpublish' : 'Publish' }}
                  </Button>
                  <Button variant="ghost" size="sm" class="text-destructive" @click="remove(a)">Delete</Button>
                </template>
              </TableCell>
            </TableRow>
            <TableRow v-if="articles.length === 0">
              <TableCell colspan="4" class="text-center text-muted-foreground">No articles yet.</TableCell>
            </TableRow>
          </TableBody>
        </Table>
      </CardContent>
    </Card>
  </div>
</template>
