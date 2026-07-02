<script setup lang="ts">
const colorMode = useColorMode()

function toggle() {
  colorMode.preference = colorMode.value === 'dark' ? 'light' : 'dark'
}
</script>

<template>
  <Button variant="ghost" size="icon" aria-label="Toggle theme" @click="toggle">
    <Icon v-if="colorMode.value === 'dark'" name="lucide:sun" class="size-4" />
    <Icon v-else name="lucide:moon" class="size-4" />
  </Button>
</template>
