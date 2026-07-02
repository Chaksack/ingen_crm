<script setup lang="ts">
defineProps<{
  links: { title: string, href: string }[]
}>()
</script>

<template>
  <Breadcrumb>
    <BreadcrumbList>
      <template v-for="(link, index) in links" :key="link.href">
        <BreadcrumbItem>
          <BreadcrumbPage v-if="index === links.length - 1">{{ link.title }}</BreadcrumbPage>
          <BreadcrumbLink v-else as-child>
            <NuxtLink :to="link.href">{{ link.title }}</NuxtLink>
          </BreadcrumbLink>
        </BreadcrumbItem>
        <BreadcrumbSeparator v-if="index < links.length - 1" />
      </template>
    </BreadcrumbList>
  </Breadcrumb>
</template>
