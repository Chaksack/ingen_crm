<script setup lang="ts">
const isSheetOpen = ref(false)
const table = ref<{ refresh: () => Promise<void> } | null>(null)

function onCreated() {
  isSheetOpen.value = false
  table.value?.refresh()
}
</script>

<template>
  <FinanceLayout>
    <div class="flex flex-wrap items-center justify-between gap-2">
      <h3 class="text-lg font-medium">
        Expenses
      </h3>
      <Sheet v-model:open="isSheetOpen">
        <SheetTrigger as-child>
          <Button>
            <Icon name="i-lucide-plus" />
            New Expense
          </Button>
        </SheetTrigger>
        <SheetContent class="w-1/2 rounded-l-lg sm:max-w-none overflow-y-auto">
          <SheetHeader>
            <SheetTitle>Record New Expense</SheetTitle>
            <SheetDescription>
              Fill in the details to record a new expense.
            </SheetDescription>
          </SheetHeader>
          <FinanceCreateExpense @created="onCreated" />
        </SheetContent>
      </Sheet>
    </div>
    <FinanceExpenseTable ref="table" />
  </FinanceLayout>
</template>
