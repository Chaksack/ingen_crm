<script setup lang="ts">
import { isNavGroup, moduleNavConfig } from '~/constants/moduleNav'

const auth = useAuthStore()

const moduleCards = computed(() =>
  (auth.manifest?.modules ?? [])
    .map((key) => moduleNavConfig[key])
    .filter((item) => !!item),
)
</script>

<template>
  <div class="flex flex-col gap-6">
    <div>
      <h1 class="text-2xl font-semibold tracking-tight">
        Welcome, {{ auth.user?.display_name }}
      </h1>
      <p class="mt-1 text-muted-foreground">
        Your organization has {{ auth.manifest?.modules.length ?? 0 }} module(s) enabled.
      </p>
    </div>

    <div class="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-3">
      <Card v-for="item in moduleCards" :key="item.title">
        <CardHeader>
          <CardTitle class="flex items-center gap-2">
            <Icon v-if="item.icon" :name="item.icon" class="size-4 text-muted-foreground" />
            {{ item.title }}
          </CardTitle>
        </CardHeader>
        <CardContent class="flex flex-wrap gap-2">
          <template v-if="isNavGroup(item)">
            <Button v-for="child in item.children" :key="child.link" as-child variant="outline" size="sm">
              <NuxtLink :to="child.link">{{ child.title }}</NuxtLink>
            </Button>
          </template>
          <Button v-else as-child variant="outline" size="sm">
            <NuxtLink :to="item.link">Open {{ item.title }}</NuxtLink>
          </Button>
        </CardContent>
      </Card>
    </div>
  </div>
</template>
