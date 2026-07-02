<script setup lang="ts">
import type { NavGroup } from '~/types/nav'
import { useSidebar } from '~/components/ui/sidebar'

const props = defineProps<{
  item: NavGroup
}>()

const { setOpenMobile } = useSidebar()
const route = useRoute()

const isActiveGroup = computed(() => props.item.children.some((c) => route.path.startsWith(c.link)))
const open = ref(isActiveGroup.value)
</script>

<template>
  <Collapsible v-model:open="open" as-child class="group/collapsible">
    <SidebarMenuItem>
      <CollapsibleTrigger as-child>
        <SidebarMenuButton :tooltip="item.title">
          <Icon v-if="item.icon" :name="item.icon" />
          <span>{{ item.title }}</span>
          <Icon
            name="lucide:chevron-right"
            class="ml-auto transition-transform duration-200 group-data-[state=open]/collapsible:rotate-90"
          />
        </SidebarMenuButton>
      </CollapsibleTrigger>
      <CollapsibleContent>
        <SidebarMenuSub>
          <SidebarMenuSubItem v-for="child in item.children" :key="child.link">
            <SidebarMenuSubButton as-child :data-active="child.link === route.path">
              <NuxtLink :to="child.link" @click="setOpenMobile(false)">
                <span>{{ child.title }}</span>
              </NuxtLink>
            </SidebarMenuSubButton>
          </SidebarMenuSubItem>
        </SidebarMenuSub>
      </CollapsibleContent>
    </SidebarMenuItem>
  </Collapsible>
</template>
