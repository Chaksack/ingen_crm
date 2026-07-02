<script setup lang="ts">
interface CollabUser { id: string; display_name: string; email: string; online: boolean }

const api = useApi()
const users = ref<CollabUser[]>([])

onMounted(async () => {
  users.value = await api<CollabUser[]>('/collab/users') ?? []
})
</script>

<template>
  <div>
    <h1 class="text-2xl font-semibold tracking-tight">Chat</h1>
    <p class="mt-1 text-muted-foreground">Pick a teammate to start a 1:1 conversation.</p>
    <div class="mt-4 max-w-md space-y-2">
      <NuxtLink
        v-for="u in users"
        :key="u.id"
        :to="`/chat/${u.id}`"
        class="flex items-center gap-3 rounded-md border border-border p-3 hover:bg-muted"
      >
        <Avatar class="size-8">
          <AvatarFallback>{{ u.display_name.slice(0, 2).toUpperCase() }}</AvatarFallback>
        </Avatar>
        <div class="flex-1">
          <div class="text-sm font-medium">{{ u.display_name }}</div>
          <div class="text-xs text-muted-foreground">{{ u.email }}</div>
        </div>
        <span
          class="size-2 rounded-full"
          :class="u.online ? 'bg-green-500' : 'bg-muted-foreground/30'"
        />
      </NuxtLink>
      <p v-if="users.length === 0" class="text-sm text-muted-foreground">
        No other teammates in your organization yet.
      </p>
    </div>
  </div>
</template>
