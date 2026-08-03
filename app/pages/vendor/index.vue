<script setup lang="ts">
const isSheetOpen = ref(false)
const table = ref<{ refresh: () => Promise<void> } | null>(null)

function onCreated() {
  isSheetOpen.value = false
  table.value?.refresh()
}
</script>

<template>
  <div class="w-full flex flex-col gap-4">
    <div class="flex flex-wrap items-center justify-between gap-2">
      <h2 class="text-2xl font-bold tracking-tight">
        Vendors
      </h2>
      <div class="flex items-center space-x-2">
        <Sheet v-model:open="isSheetOpen">
          <SheetTrigger as-child>
            <Button>
              <Icon name="i-lucide-plus" class="" />
              New Vendors
            </Button>
          </SheetTrigger>
          <SheetContent class="w-1/2 rounded-l-lg sm:max-w-none overflow-y-auto ">
            <SheetHeader>
              <SheetTitle>Create New Vendor</SheetTitle>
              <SheetDescription>
                Fill in the details to create a new vendor.
              </SheetDescription>
            </SheetHeader>
            <VendorCreate @created="onCreated" />
          </SheetContent>
        </Sheet>
      </div>
    </div>
    <main>
      <VendorTable ref="table" />
    </main>
  </div>
</template>
