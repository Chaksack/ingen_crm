<script setup lang="ts">
definePageMeta({ layout: 'auth' })

const organizationName = ref('')
const displayName = ref('')
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
    const res = await api('/auth/register', {
      method: 'POST',
      body: {
        organization_name: organizationName.value,
        display_name: displayName.value,
        email: email.value,
        password: password.value,
      },
    })
    auth.setToken(res.access_token)
    await navigateTo('/')
  } catch {
    error.value = 'Could not create the organization. Check your details and try again.'
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <Card class="w-full max-w-sm">
    <CardHeader>
      <CardTitle>Create your organization</CardTitle>
      <CardDescription>Sets you up as the Admin with Sales, Service, and Chat enabled.</CardDescription>
    </CardHeader>
    <CardContent>
      <form class="space-y-4" @submit.prevent="onSubmit">
        <div class="space-y-1.5">
          <Label for="org">Organization name</Label>
          <Input id="org" v-model="organizationName" required />
        </div>
        <div class="space-y-1.5">
          <Label for="name">Your name</Label>
          <Input id="name" v-model="displayName" required />
        </div>
        <div class="space-y-1.5">
          <Label for="email">Email</Label>
          <Input id="email" v-model="email" type="email" autocomplete="email" required />
        </div>
        <div class="space-y-1.5">
          <Label for="password">Password</Label>
          <Input id="password" v-model="password" type="password" autocomplete="new-password" minlength="8" required />
        </div>
        <p v-if="error" class="text-sm text-destructive">{{ error }}</p>
        <Button type="submit" class="w-full" :disabled="loading">
          {{ loading ? 'Creating…' : 'Create organization' }}
        </Button>
      </form>
      <p class="mt-4 text-center text-sm text-muted-foreground">
        Already have an account?
        <NuxtLink to="/login" class="text-primary underline-offset-4 hover:underline">Sign in</NuxtLink>
      </p>
    </CardContent>
  </Card>
</template>
