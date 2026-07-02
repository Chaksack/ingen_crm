<script setup lang="ts">
const chat = useChatStore()

onMounted(() => {
  chat.connect()
})
</script>

<template>
  <SidebarProvider>
    <LayoutAppSidebar />
    <SidebarInset>
      <LayoutHeader />
      <div class="flex flex-1 flex-col">
        <div class="grow p-4 lg:p-6">
          <slot />
        </div>
      </div>
    </SidebarInset>
  </SidebarProvider>
</template>
