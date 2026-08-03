<script lang="ts" setup>
import { toast } from 'vue-sonner'

const otp = ref<string[]>([])
const isLoading = ref(false)
const isResending = ref(false)
const { verifyOtp, resendOtp, fetchUser } = useAuth()

async function onSubmit(event: Event) {
  event.preventDefault()
  const code = otp.value.join('')
  if (code.length < 6)
    return

  isLoading.value = true
  try {
    const { purpose } = await verifyOtp(code)
    if (purpose === 'login') {
      await fetchUser()
      await navigateTo('/dashboard')
    }
    else {
      await navigateTo('/new-password')
    }
  }
  catch (err: any) {
    toast.error(err?.data?.statusMessage || 'Invalid or expired code')
  }
  finally {
    isLoading.value = false
  }
}

async function onResend() {
  isResending.value = true
  try {
    await resendOtp()
    toast.success('A new code has been sent to your email.')
  }
  catch (err: any) {
    toast.error(err?.data?.statusMessage || 'Could not resend code')
  }
  finally {
    isResending.value = false
  }
}
</script>

<template>
  <div class="flex flex-col gap-6">
    <form @submit="onSubmit">
      <FieldGroup>
        <div class="flex flex-col items-center gap-1 text-center">
          <h1 class="text-2xl font-bold">
            Enter verification code
          </h1>
          <p class="text-muted-foreground text-sm text-balance">
            We sent a 6-digit code to your email.
          </p>
        </div>
        <Field>
          <FieldLabel html-for="otp" class="sr-only">
            Verification code
          </FieldLabel>
          <PinInput id="otp" v-model="otp" class="justify-center">
            <PinInputGroup class="gap-1 *:data-[slot=pin-input-slot]:rounded-md *:data-[slot=pin-input-slot]:border">
              <template v-for="(id, index) in 6" :key="id">
                <PinInputSlot :index="index" />
                <template v-if="index !== 5">
                  <PinInputSeparator />
                </template>
              </template>
            </PinInputGroup>
          </PinInput>
          <FieldDescription class="text-center">
            Enter the 6-digit code sent to your email.
          </FieldDescription>
        </Field>
        <Button type="submit" :disabled="isLoading">
          Verify
        </Button>
        <FieldDescription class="text-center">
          Didn&apos;t receive the code?
          <a href="#" @click.prevent="!isResending && onResend()">Resend</a>
        </FieldDescription>
      </FieldGroup>
    </form>
  </div>
</template>

<style>

</style>
