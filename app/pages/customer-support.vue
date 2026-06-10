<script setup lang="ts">
import { ref } from 'vue'
import { columns } from '@/components/tasks/components/columns'
import DataTable from '@/components/tasks/components/DataTable.vue'
import tasksData from '@/components/tasks/data/tasks.json'
import type { Task } from '@/components/tasks/data/schema'
import SupportChatSheet from '@/components/Customer/SupportChatSheet.vue'

const tasks = ref(tasksData.data)

const isSheetOpen = ref(false)
const selectedTask = ref<Task | null>(null)

function handleRowClick(task: Task) {
  selectedTask.value = task
  isSheetOpen.value = true
}

function handleSheetClose(value: boolean) {
  isSheetOpen.value = value
  if (!value) {
    selectedTask.value = null
  }
}

function handleTaskUpdate(updatedTask: Task) {
  const index = tasks.value.findIndex(t => t.id === updatedTask.id)
  if (index !== -1) {
    tasks.value.splice(index, 1, updatedTask)
  }
}
</script>

<template>
  <div class="w-full flex flex-col items-stretch gap-4">
    <div class="flex flex-wrap items-end justify-between gap-2">
      <div>
        <h2 class="text-2xl font-bold tracking-tight">
          Customer Support
        </h2>
        <p class="text-muted-foreground">
          Here&apos;s a list of customer support tasks!
        </p>
      </div>
    </div>
    <DataTable :data="tasks" :columns="columns" @row-click="handleRowClick" />
    <SupportChatSheet
      :open="isSheetOpen"
      :task="selectedTask"
      @update:open="handleSheetClose"
      @update:task="handleTaskUpdate"
    />
  </div>
</template>

<style scoped>

</style>
