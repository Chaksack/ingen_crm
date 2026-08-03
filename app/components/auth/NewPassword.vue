<script setup lang="ts">
import { Loader2 } from 'lucide-vue-next'
import { toast } from 'vue-sonner'

const isLoading = ref(false)
const password = ref('')
const confirmPassword = ref('')
const { resetPassword } = useAuth()

async function onSubmit(event: Event) {
  event.preventDefault()
  if (password.value.length < 6) {
    toast.error('Password must be at least 6 characters')
    return
  }
  if (password.value !== confirmPassword.value) {
    toast.error('Passwords do not match')
    return
  }

  isLoading.value = true
  try {
    await resetPassword(password.value)
    toast.success('Password updated. Please sign in.')
    await navigateTo('/')
  }
  catch (err: any) {
    toast.error(err?.data?.statusMessage || 'Could not reset password')
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
