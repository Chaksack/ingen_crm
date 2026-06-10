<script setup lang="ts">
definePageMeta({
  layout: 'blank',
})
</script>

<template>
  <div class="flex flex-col items-center justify-center gap-6 bg-muted p-6 min-h-svh md:p-10">
    <div class="max-w-sm w-full flex flex-col gap-6">
      <NuxtLink to="#" class="flex items-center self-center gap-2 font-medium">
        <img src="@/assets/logo/logo-2-light.png" alt="Logo" class="h-18 hidden dark:block" />
        <img src="@/assets/logo/logo-2.png" alt="Logo" class="h-18  dark:hidden" />

      </NuxtLink>
      <div class="flex flex-col gap-6">
        <Card>
          <CardHeader class="text-center">
            <CardTitle class="text-xl">
              Welcome to Ingen Cloud Technologies
            </CardTitle>
            <CardDescription>
              Login to your staff account to continue
            </CardDescription>
          </CardHeader>
          <CardContent>
            <AuthSignIn />
          </CardContent>
        </Card>
        <div class="text-center text-balance text-xs text-muted-foreground [&_a]:underline [&_a]:underline-offset-4 [&_a]:hover:text-primary">
          © 2026 Ingen Cloud Technologies. All rights reserved.
        </div>
      </div>
    </div>
  </div>
</template>
