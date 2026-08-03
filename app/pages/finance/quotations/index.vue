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
        Quotations
      </h3>
      <Sheet v-model:open="isSheetOpen">
        <SheetTrigger as-child>
          <Button>
            <Icon name="i-lucide-plus" />
            New Quotation
          </Button>
        </SheetTrigger>
        <SheetContent class="w-1/2 rounded-l-lg sm:max-w-none overflow-y-auto">
          <SheetHeader>
            <SheetTitle>Create New Quotation</SheetTitle>
            <SheetDescription>
              Add line items and generate a new quotation.
            </SheetDescription>
          </SheetHeader>
          <FinanceCreateQuotation @created="onCreated" />
        </SheetContent>
      </Sheet>
    </div>
    <FinanceQuotationTable ref="table" />
  </FinanceLayout>
</template>
