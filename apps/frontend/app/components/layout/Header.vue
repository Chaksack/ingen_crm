<script setup lang="ts">
const route = useRoute()
const breadcrumbOverride = useBreadcrumbOverride()

function toTitle(segment: string) {
  return segment
    .replace(/-/g, ' ')
    .split(' ')
    .map((w) => w.charAt(0).toUpperCase() + w.slice(1))
    .join(' ')
}

const links = computed(() => {
  if (route.path === '/') return [{ title: 'Home', href: '/' }]
  const segments = route.path.split('/').filter(Boolean)
  const crumbs = segments.map((segment, index) => ({
    title: index === segments.length - 1 && breadcrumbOverride.value ? breadcrumbOverride.value : toTitle(segment),
    href: `/${segments.slice(0, index + 1).join('/')}`,
  }))
  return [{ title: 'Home', href: '/' }, ...crumbs]
})

watch(() => route.path, () => {
  breadcrumbOverride.value = null
})
</script>

<template>
  <header class="sticky top-0 z-10 flex h-14 items-center gap-4 border-b border-border bg-background px-4 md:px-6">
    <div class="flex h-4 w-full items-center gap-4">
      <SidebarTrigger />
      <Separator orientation="vertical" />
      <BaseBreadcrumbCustom :links="links" />
    </div>
    <div class="ml-auto flex items-center gap-2">
      <LayoutNotificationBell />
      <DarkToggle />
    </div>
  </header>
</template>
