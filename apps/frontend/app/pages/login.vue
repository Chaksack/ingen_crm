<script setup lang="ts">
definePageMeta({ layout: 'auth' })

const email = ref('')
const password = ref('')
const error = ref('')
const loading = ref(false)

const auth = useAuthStore()
const api = useApi()

async function onSubmit() {
  error.value = ''
  loading.value = true
  try {
    const res = await api('/auth/login', {
      method: 'POST',
      body: { email: email.value, password: password.value },
    })
    auth.setToken(res.access_token)
    await navigateTo('/')
  } catch {
    error.value = 'Invalid email or password.'
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <Card class="w-full max-w-sm">
    <CardHeader>
      <CardTitle>Sign in to Ingen One</CardTitle>
      <CardDescription>Use your organization credentials.</CardDescription>
    </CardHeader>
    <CardContent>
      <form class="space-y-4" @submit.prevent="onSubmit">
        <div class="space-y-1.5">
          <Label for="email">Email</Label>
          <Input id="email" v-model="email" type="email" autocomplete="email" required />
        </div>
        <div class="space-y-1.5">
          <Label for="password">Password</Label>
          <Input id="password" v-model="password" type="password" autocomplete="current-password" required />
        </div>
        <p v-if="error" class="text-sm text-destructive">{{ error }}</p>
        <Button type="submit" class="w-full" :disabled="loading">
          {{ loading ? 'Signing in…' : 'Sign in' }}
        </Button>
      </form>
      <p class="mt-4 text-center text-sm text-muted-foreground">
        No organization yet?
        <NuxtLink to="/register" class="text-primary underline-offset-4 hover:underline">Create one</NuxtLink>
      </p>
    </CardContent>
  </Card>
</template>
