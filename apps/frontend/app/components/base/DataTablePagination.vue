<script setup lang="ts">
import type { Table } from '@tanstack/vue-table'

defineProps<{
  table: Table<any>
}>()
</script>

<template>
  <div class="flex items-center justify-between">
    <div class="flex-1 text-sm text-muted-foreground">
      {{ table.getFilteredRowModel().rows.length }} row(s)
    </div>
    <div class="flex items-center space-x-6 lg:space-x-8">
      <div class="flex items-center space-x-2">
        <p class="text-sm font-medium">Rows per page</p>
        <Select :model-value="`${table.getState().pagination.pageSize}`" @update:model-value="(v) => table.setPageSize(Number(v))">
          <SelectTrigger class="h-8 w-[70px]">
            <SelectValue :placeholder="`${table.getState().pagination.pageSize}`" />
          </SelectTrigger>
          <SelectContent side="top">
            <SelectItem v-for="pageSize in [10, 20, 30, 50]" :key="pageSize" :value="`${pageSize}`">
              {{ pageSize }}
            </SelectItem>
          </SelectContent>
        </Select>
      </div>
      <div class="flex w-[100px] items-center justify-center text-sm font-medium">
        Page {{ table.getState().pagination.pageIndex + 1 }} of {{ table.getPageCount() || 1 }}
      </div>
      <div class="flex items-center space-x-2">
        <Button variant="outline" size="icon" :disabled="!table.getCanPreviousPage()" @click="table.previousPage()">
          <Icon name="lucide:chevron-left" class="size-4" />
        </Button>
        <Button variant="outline" size="icon" :disabled="!table.getCanNextPage()" @click="table.nextPage()">
          <Icon name="lucide:chevron-right" class="size-4" />
        </Button>
      </div>
    </div>
  </div>
</template>
