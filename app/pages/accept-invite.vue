<script setup lang="ts">
import { Loader2 } from 'lucide-vue-next'
import { toast } from 'vue-sonner'

definePageMeta({
  layout: 'blank',
})

const route = useRoute()
const token = computed(() => String(route.query.token ?? ''))

const isChecking = ref(true)
const inviteValid = ref(false)
const inviteError = ref('')
const inviteEmail = ref('')

const isSubmitting = ref(false)
const done = ref(false)
const password = ref('')
const confirmPassword = ref('')

onMounted(async () => {
  if (!token.value) {
    inviteError.value = 'This invite link is missing a token.'
    isChecking.value = false
    return
  }
  try {
    const info = await $fetch<{ email: string, name: string }>('/api/auth/invite-info', {
      query: { token: token.value },
    })
    inviteEmail.value = info.email
    inviteValid.value = true
  }
  catch (err: any) {
    inviteError.value = err?.data?.statusMessage || 'This invite link is invalid or has expired.'
  }
  finally {
    isChecking.value = false
  }
})

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

  isSubmitting.value = true
  try {
    await $fetch('/api/auth/accept-invite', {
      method: 'POST',
      body: { token: token.value, password: password.value },
    })
    done.value = true
  }
  catch (err: any) {
    toast.error(err?.data?.statusMessage || 'Could not activate your account')
  }
  finally {
    isSubmitting.value = false
  }
}
</script>

<template>
  <div class="flex flex-col items-center justify-center gap-6 bg-muted p-6 min-h-svh md:p-10">
    <div class="max-w-sm w-full flex flex-col gap-6">
      <h1 class="font-bold text-2xl text-center">
        Ingenicx
      </h1>

      <Card v-if="isChecking">
        <CardContent class="pt-6 text-center text-sm text-muted-foreground">
          Checking your invite...
        </CardContent>
      </Card>

      <Card v-else-if="done">
        <CardHeader class="text-center">
          <CardTitle class="text-xl">
            You're all set
          </CardTitle>
          <CardDescription>
            Your password has been created. You can now sign in.
          </CardDescription>
        </CardHeader>
        <CardContent>
          <NuxtLink to="/">
            <Button class="w-full">
              Go to Login
            </Button>
          </NuxtLink>
        </CardContent>
      </Card>

      <Card v-else-if="!inviteValid">
        <CardHeader class="text-center">
          <CardTitle class="text-xl">
            Invite not valid
          </CardTitle>
          <CardDescription>
            {{ inviteError }}
          </CardDescription>
        </CardHeader>
      </Card>

      <Card v-else>
        <CardHeader class="text-center">
          <CardTitle class="text-xl">
            Welcome to Ingenicx
          </CardTitle>
          <CardDescription>
            Set a password for {{ inviteEmail }}
          </CardDescription>
        </CardHeader>
        <CardContent>
          <form class="grid gap-4" @submit="onSubmit">
            <div class="grid gap-2">
              <Label for="password">Password</Label>
              <PasswordInput id="password" v-model="password" :disabled="isSubmitting" />
            </div>
            <div class="grid gap-2">
              <Label for="confirmPassword">Confirm Password</Label>
              <PasswordInput id="confirmPassword" v-model="confirmPassword" :disabled="isSubmitting" />
            </div>
            <Button type="submit" class="w-full" :disabled="isSubmitting">
              <Loader2 v-if="isSubmitting" class="mr-2 h-4 w-4 animate-spin" />
              Activate Account
            </Button>
          </form>
        </CardContent>
      </Card>
    </div>
  </div>
</template>
