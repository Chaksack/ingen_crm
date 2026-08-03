<script setup lang="ts">
import { useRoute, useRouter } from 'vue-router'
import ProjectGantt from '~/components/projects/ProjectGantt.vue'
import { useProjects } from '~/composables/useProjects'

const route = useRoute()
const router = useRouter()
const { getProject, updateProject } = useProjects()

const pid = computed(() => String(route.params.id))
const project = computed(() => getProject(pid.value))

// Boards handling
const selectedBoard = ref<string>('Board')
watchEffect(() => {
  if (project.value) {
    const boards = project.value.boards && project.value.boards.length ? project.value.boards : ['Board']
    if (!boards.includes(selectedBoard.value))
      selectedBoard.value = boards[0]
  }
  else {
    router.replace('/projects')
  }
})

const newBoardName = ref('')
function createBoard() {
  const name = newBoardName.value.trim()
  if (!project.value || !name)
    return
  const boards = Array.from(new Set([...(project.value.boards || ['Board']), name]))
  updateProject(project.value.id, { boards })
  selectedBoard.value = name
  newBoardName.value = ''
}
</script>

<template>
  <div class="-m-4 lg:-m-6 h-full flex flex-col">
    <!-- Header -->
    <div class="flex h-[56px] items-center gap-2 border-b px-4 justify-between">
      <div class="flex items-center gap-3 min-w-0">
        <Button size="icon" variant="ghost" title="Back" @click="router.push(`/projects/${pid}/board`)">
          <Icon name="i-lucide-arrow-left" />
        </Button>
        <h2 class="text-xl font-semibold truncate">
          {{ project?.name || 'Project' }}
        </h2>
        <Badge v-if="project?.key" variant="secondary">
          {{ project?.key }}
        </Badge>
        <span v-if="project?.customerName" class="text-sm text-muted-foreground truncate">• {{ project?.customerName }}</span>
      </div>
      <div class="flex items-center gap-2">
        <div class="flex items-center gap-2">
          <Label class="sr-only">Board</Label>
          <Select v-model="selectedBoard">
            <SelectTrigger class="w-[160px]">
              <SelectValue placeholder="Select board" />
            </SelectTrigger>
            <SelectContent>
              <SelectItem v-for="b in (project?.boards || ['Board'])" :key="b" :value="b">
                {{ b }}
              </SelectItem>
            </SelectContent>
          </Select>
          <Dialog>
            <DialogTrigger as-child>
              <Button size="icon-sm" variant="outline" title="Add board">
                <Icon name="i-lucide-plus" />
              </Button>
            </DialogTrigger>
            <DialogContent class="sm:max-w-[420px]">
              <DialogHeader>
                <DialogTitle>New Board</DialogTitle>
              </DialogHeader>
              <div class="space-y-3 py-2">
                <div class="space-y-2">
                  <Label for="bname">Board name</Label>
                  <Input id="bname" v-model="newBoardName" placeholder="e.g. Development" />
                </div>
                <div class="flex items-center gap-2">
                  <DialogClose as-child>
                    <Button variant="secondary">
                      Cancel
                    </Button>
                  </DialogClose>
                  <DialogClose as-child>
                    <Button :disabled="!newBoardName.trim()" @click="createBoard">
                      Create
                    </Button>
                  </DialogClose>
                </div>
              </div>
            </DialogContent>
          </Dialog>
        </div>
        <Button size="sm" variant="outline" @click="router.push(`/projects/${pid}/board`)">
          <Icon name="lucide:layout-dashboard" /> Board
        </Button>
      </div>
    </div>

    <!-- Gantt content -->
    <section class="flex-1 min-h-0 overflow-hidden p-4">
      <ProjectGantt :project-id="pid" :board-id="selectedBoard" />
    </section>
  </div>
</template>
