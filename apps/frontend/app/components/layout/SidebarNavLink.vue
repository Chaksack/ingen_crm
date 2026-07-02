<script setup lang="ts">
import type { NavLink } from '~/types/nav'
import { useSidebar } from '~/components/ui/sidebar'

defineProps<{
  item: NavLink
}>()

const { setOpenMobile } = useSidebar()
const route = useRoute()
</script>

<template>
  <SidebarMenuItem>
    <SidebarMenuButton as-child :tooltip="item.title" :data-active="item.link === route.path">
      <NuxtLink :to="item.link" @click="setOpenMobile(false)">
        <Icon v-if="item.icon" :name="item.icon" />
        <span>{{ item.title }}</span>
      </NuxtLink>
    </SidebarMenuButton>
  </SidebarMenuItem>
</template>
