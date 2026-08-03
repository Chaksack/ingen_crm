<script setup lang="ts">
import { Loader2 } from 'lucide-vue-next'
import { toast } from 'vue-sonner'

const email = ref('')
const isLoading = ref(false)
const { forgotPassword } = useAuth()

async function onSubmit(event: Event) {
  event.preventDefault()
  if (!email.value)
    return

  isLoading.value = true
  try {
    await forgotPassword(email.value)
    toast.success('If that email exists, a reset code has been sent.')
    await navigateTo('/otp')
  }
  catch (err: any) {
    toast.error(err?.data?.statusMessage || 'Something went wrong')
  }
  finally {
    isLoading.value = false
  }
}
</script>

<template>
  <form @submit="onSubmit">
    <div class="grid gap-4">
      <div class="grid gap-2">
        <Label for="email">
          Email
        </Label>
        <Input
          id="email"
          v-model="email"
          placeholder="name@example.com"
          type="email"
          auto-capitalize="none"
          auto-complete="email"
          auto-correct="off"
          :disabled="isLoading"
        />
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
