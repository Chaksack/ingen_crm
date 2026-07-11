<script setup lang="ts">
definePageMeta({ layout: 'auth' })

const route = useRoute()
const token = computed(() => route.query.token as string | undefined)

const api = useApi()
const auth = useAuthStore()

const state = ref<'loading' | 'ready' | 'invalid' | 'submitting'>('loading')
const email = ref('')
const organizationName = ref('')
const displayName = ref('')
const password = ref('')
const error = ref('')

onMounted(async () => {
  if (!token.value) {
    state.value = 'invalid'
    return
  }
  try {
    const res = await api<{ email: string, organization_name: string }>(`/invites/${token.value}`)
    email.value = res.email
    organizationName.value = res.organization_name
    state.value = 'ready'
  } catch {
    state.value = 'invalid'
  }
})

async function onSubmit() {
  error.value = ''
  state.value = 'submitting'
  try {
    const res = await api(`/invites/${token.value}/accept`, {
      method: 'POST',
      body: { display_name: displayName.value, password: password.value },
    })
    auth.setToken(res.access_token)
    await navigateTo('/')
  } catch {
    error.value = 'Could not accept this invite. It may have already been used.'
    state.value = 'ready'
  }
}
</script>

<template>
  <Card class="w-full max-w-sm">
    <template v-if="state === 'loading'">
      <CardHeader>
        <CardTitle>Loading invite…</CardTitle>
      </CardHeader>
    </template>

    <template v-else-if="state === 'invalid'">
      <CardHeader>
        <CardTitle>Invite not found</CardTitle>
        <CardDescription>This invite link is invalid, already used, or has expired.</CardDescription>
      </CardHeader>
      <CardContent>
        <Button as-child class="w-full" variant="outline">
          <NuxtLink to="/login">Back to sign in</NuxtLink>
        </Button>
      </CardContent>
    </template>

    <template v-else>
      <CardHeader>
        <CardTitle>Join {{ organizationName }}</CardTitle>
        <CardDescription>Set your name and password for {{ email }}.</CardDescription>
      </CardHeader>
      <CardContent>
        <form class="space-y-4" @submit.prevent="onSubmit">
          <div class="space-y-1.5">
            <Label for="name">Your name</Label>
            <Input id="name" v-model="displayName" required />
          </div>
          <div class="space-y-1.5">
            <Label for="password">Password</Label>
            <Input id="password" v-model="password" type="password" autocomplete="new-password" minlength="8" required />
          </div>
          <p v-if="error" class="text-sm text-destructive">{{ error }}</p>
          <Button type="submit" class="w-full" :disabled="state === 'submitting'">
            {{ state === 'submitting' ? 'Joining…' : 'Join organization' }}
          </Button>
        </form>
      </CardContent>
    </template>
  </Card>
</template>
