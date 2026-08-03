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
        Invoices
      </h3>
      <Sheet v-model:open="isSheetOpen">
        <SheetTrigger as-child>
          <Button>
            <Icon name="i-lucide-plus" />
            New Invoice
          </Button>
        </SheetTrigger>
        <SheetContent class="w-1/2 rounded-l-lg sm:max-w-none overflow-y-auto">
          <SheetHeader>
            <SheetTitle>Create New Invoice</SheetTitle>
            <SheetDescription>
              Add line items and generate a new invoice.
            </SheetDescription>
          </SheetHeader>
          <FinanceCreateInvoice @created="onCreated" />
        </SheetContent>
      </Sheet>
    </div>
    <FinanceInvoiceTable ref="table" />
  </FinanceLayout>
</template>
