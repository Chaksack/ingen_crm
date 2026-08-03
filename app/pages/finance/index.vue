<script setup lang="ts">

</script>

<template>
  <div class="w-full flex flex-col gap-4">
    <div class="flex flex-wrap items-center justify-between gap-2">
      <h2 class="text-2xl font-bold tracking-tight">
        Financial Institutions
      </h2>
     <div class="flex items-center space-x-2">
       <Sheet>
    <SheetTrigger as-child>
        <Button>
          <Icon name="i-lucide-plus" class="" />
          New Financial Institution</Button>
        </SheetTrigger>
        <SheetContent class="w-1/2 rounded-l-lg sm:max-w-none overflow-y-auto ">
          <SheetHeader>
            <SheetTitle>Create New Financial Institution</SheetTitle>
            <SheetDescription>
              Fill in the details to create a new financial institution.
            </SheetDescription>
          </SheetHeader>
          <FICreateFinancialInstitution />
        </SheetContent>
      </Sheet>
      </div>
    </div>
    <main>
      <FIFinancialInstitutionTable />
    </main>
  </div>
</template>