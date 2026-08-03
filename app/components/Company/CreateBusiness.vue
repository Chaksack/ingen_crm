<script setup lang="ts">
import { Loader2 } from 'lucide-vue-next'
import { toast } from 'vue-sonner'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'

const emit = defineEmits<{ created: [] }>()

const isLoading = ref(false)
const form = reactive({
  companyName: '',
  companyNumber: '',
  email: '',
  phoneNumber: '',
  address: '',
  capital: '',
  foundedYear: '',
  country: '',
  currency: 'GHS',
  status: 'pending' as 'active' | 'inactive' | 'pending',
})

async function onSubmit(event: Event) {
  event.preventDefault()
  if (!form.companyName || !form.email) {
    toast.error('Company name and email are required')
    return
  }

  isLoading.value = true
  try {
    await $fetch('/api/businesses', {
      method: 'POST',
      body: {
        ...form,
        capital: form.capital ? Number(form.capital) : undefined,
        foundedYear: form.foundedYear ? Number(form.foundedYear) : undefined,
      },
    })
    toast.success('Business created')
    emit('created')
  }
  catch (err: any) {
    toast.error(err?.data?.statusMessage || 'Failed to create business')
  }
  finally {
    isLoading.value = false
  }
}
</script>

<template>
  <form class="grid gap-4 px-4" @submit="onSubmit">
    <div class="grid gap-2">
      <Label for="companyName">Company Name</Label>
      <Input id="companyName" v-model="form.companyName" :disabled="isLoading" />
    </div>
    <div class="grid gap-2">
      <Label for="email">Email</Label>
      <Input id="email" v-model="form.email" type="email" :disabled="isLoading" />
    </div>
    <div class="grid grid-cols-2 gap-4">
      <div class="grid gap-2">
        <Label for="companyNumber">Company Number</Label>
        <Input id="companyNumber" v-model="form.companyNumber" :disabled="isLoading" />
      </div>
      <div class="grid gap-2">
        <Label for="phoneNumber">Phone Number</Label>
        <Input id="phoneNumber" v-model="form.phoneNumber" :disabled="isLoading" />
      </div>
    </div>
    <div class="grid gap-2">
      <Label for="address">Address</Label>
      <Input id="address" v-model="form.address" :disabled="isLoading" />
    </div>
    <div class="grid grid-cols-2 gap-4">
      <div class="grid gap-2">
        <Label for="capital">Capital</Label>
        <Input id="capital" v-model="form.capital" type="number" min="0" step="0.01" :disabled="isLoading" />
      </div>
      <div class="grid gap-2">
        <Label for="foundedYear">Founded Year</Label>
        <Input id="foundedYear" v-model="form.foundedYear" type="number" :disabled="isLoading" />
      </div>
    </div>
    <div class="grid grid-cols-2 gap-4">
      <div class="grid gap-2">
        <Label for="country">Country</Label>
        <Input id="country" v-model="form.country" :disabled="isLoading" />
      </div>
      <div class="grid gap-2">
        <Label for="currency">Currency</Label>
        <Input id="currency" v-model="form.currency" :disabled="isLoading" />
      </div>
    </div>
    <div class="grid gap-2">
      <Label for="status">Status</Label>
      <Select v-model="form.status">
        <SelectTrigger id="status" class="w-full">
          <SelectValue />
        </SelectTrigger>
        <SelectContent>
          <SelectItem value="active">
            Active
          </SelectItem>
          <SelectItem value="inactive">
            Inactive
          </SelectItem>
          <SelectItem value="pending">
            Pending
          </SelectItem>
        </SelectContent>
      </Select>
    </div>
    <Button type="submit" :disabled="isLoading" class="mt-2">
      <Loader2 v-if="isLoading" class="mr-2 h-4 w-4 animate-spin" />
      Create Business
    </Button>
  </form>
</template>
