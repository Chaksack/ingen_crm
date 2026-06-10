<script setup lang="ts">
import { Loader2 } from 'lucide-vue-next'

const isLoading = ref(false)
const password = ref('')
const confirmPassword = ref('')

async function onSubmit(event: Event) {
  event.preventDefault()
  isLoading.value = true

  setTimeout(() => {
    isLoading.value = false
  }, 3000)
}
</script>

<template>
  <form @submit="onSubmit">
    <div class="grid gap-4">
      <div class="grid gap-2">
          <Label for="password">
          Password
        </Label>
      <PasswordInput id="password" v-model="password" />
      </div>
       <div class="grid gap-2">
          <Label for="confirm-password">
          Confirm Password
        </Label>
      <PasswordInput id="confirm-password" v-model="confirmPassword" />
      </div>
      <Button :disabled="isLoading">
        <Loader2 v-if="isLoading" class="mr-2 h-4 w-4 animate-spin" />
        Submit
      </Button>
    </div>
  </form>
</template>

<style scoped>

</style>
