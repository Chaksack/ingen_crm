<script setup lang="ts">
import { isNavGroup, moduleNavConfig } from '~/constants/moduleNav'
import type { NavItem } from '~/types/nav'

const auth = useAuthStore()

const navItems = computed<NavItem[]>(() =>
  (auth.manifest?.modules ?? [])
    .map((key) => moduleNavConfig[key])
    .filter((item): item is NavItem => !!item),
)

function resolveNavItemComponent(item: NavItem) {
  return isNavGroup(item) ? resolveComponent('LayoutSidebarNavGroup') : resolveComponent('LayoutSidebarNavLink')
}
</script>

<template>
  <Sidebar collapsible="icon">
    <SidebarHeader>
      <SidebarMenu>
        <SidebarMenuItem>
          <SidebarMenuButton size="lg" class="cursor-default hover:bg-transparent">
            <div class="flex aspect-square size-8 items-center justify-center rounded-lg bg-sidebar-primary text-sidebar-primary-foreground">
              <Icon name="lucide:layout-grid" class="size-4" />
            </div>
            <div class="grid flex-1 text-left text-sm leading-tight">
              <span class="truncate font-semibold">Ingen One</span>
              <span class="truncate text-xs text-muted-foreground">{{ auth.user?.organization_id ? 'Workspace' : '' }}</span>
            </div>
          </SidebarMenuButton>
        </SidebarMenuItem>
      </SidebarMenu>
    </SidebarHeader>
    <SidebarContent>
      <SidebarGroup>
        <SidebarGroupLabel>General</SidebarGroupLabel>
        <SidebarMenu>
          <LayoutSidebarNavLink :item="{ title: 'Home', link: '/', icon: 'lucide:home' }" />
        </SidebarMenu>
      </SidebarGroup>
      <SidebarGroup v-if="navItems.length">
        <SidebarGroupLabel>Modules</SidebarGroupLabel>
        <SidebarMenu>
          <component :is="resolveNavItemComponent(item)" v-for="(item, index) in navItems" :key="index" :item="item" />
        </SidebarMenu>
      </SidebarGroup>
    </SidebarContent>
    <SidebarFooter>
      <LayoutSidebarNavFooter />
    </SidebarFooter>
    <SidebarRail />
  </Sidebar>
</template>
