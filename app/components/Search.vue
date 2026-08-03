<script setup lang="ts">
import type { NavGroup, NavMenu } from '~/types/nav'
import { navMenu } from '@/constants/menus'

const { metaSymbol } = useShortcuts()
const router = useRouter()
const openCommand = ref(false)

defineShortcuts({
  Meta_K: () => openCommand.value = true,
})

const componentsNav = computed<NavGroup | undefined>(() => {
  return navMenu
    .flatMap((nav: NavMenu) => nav.items)
    .find((item: NavGroup) => item.title === 'Companies')
})

function handleSelectCustomer(id: string) {
  router.push(`/customer/view/${id}`)
  openCommand.value = false
}

function handleSelectCompany(id: string) {
  router.push(`/company/view/${id}`)
  openCommand.value = false
}
</script>

<template>
  <div as-child tooltip="Search">
    <Button 
      variant="outline" 
      size="sm" 
      class="text-xs" 
      @click="openCommand = !openCommand"
    >
      <Icon name="i-lucide-search" />
      <span class="font-normal group-data-[collapsible=icon]:hidden">
        Search for User/Company
      </span>
      <div class="ml-auto flex items-center space-x-0.5 group-data-[collapsible=icon]:hidden">
        <Kbd>{{ metaSymbol }}</Kbd>
        <Kbd>K</Kbd>
      </div>
    </Button>
  </div>

  <CommandDialog v-model:open="openCommand">
    <CommandInput placeholder="Type a user's name or GhcardID/Business Registration Number..." />
    <CommandList>
      <CommandEmpty>No results found.</CommandEmpty>
      
      <CommandGroup heading="Customers">
        <CommandItem value="john-snow" @select="handleSelectCustomer('1')">
          John Snow 
        </CommandItem>
        <CommandItem value="william-apnoch" @select="handleSelectCustomer('2')">
          William Apnoch
        </CommandItem>
      </CommandGroup>
      
      <CommandSeparator />
      
      <CommandGroup heading="Companies">
        <CommandItem value="acme-corporation" @select="handleSelectCompany('1')">
          ACME Corporation 
        </CommandItem>
        <CommandItem value="globex-inc" @select="handleSelectCompany('2')">
          Globex Inc.
        </CommandItem>
      </CommandGroup>
    </CommandList>
  </CommandDialog>
</template>