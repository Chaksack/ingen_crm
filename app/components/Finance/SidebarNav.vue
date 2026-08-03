<script setup lang="ts">
import { cn } from '@/lib/utils'

interface Item {
  title: string
  href: string
}

const route = useRoute()

const sidebarNavItems: Item[] = [
  { title: 'Overview', href: '/finance' },
  { title: 'Documents', href: '/finance/documents' },
  { title: 'Payments', href: '/finance/payments' },
  { title: 'Expenses', href: '/finance/expenses' },
  { title: 'Tax Settings', href: '/finance/tax-settings' },
]
</script>

<template>
  <nav class="flex lg:flex-col space-x-2 lg:space-x-0 lg:space-y-1">
    <Button
      v-for="item in sidebarNavItems"
      :key="item.title"
      variant="ghost"
      :class="cn(
        'w-full text-left justify-start items-start',
        route.path === item.href && 'bg-muted hover:bg-muted',
      )"
      as-child
    >
      <NuxtLink :to="item.href">
        {{ item.title }}
      </NuxtLink>
    </Button>
  </nav>
</template>
