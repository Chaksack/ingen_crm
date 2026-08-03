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
import { Textarea } from '@/components/ui/textarea'

const emit = defineEmits<{ created: [] }>()

const isLoading = ref(false)
const form = reactive({
  vendorName: '',
  email: '',
  phone: '',
  address: '',
  website: '',
  vendorType: '',
  registrationNumber: '',
  foundedDate: '',
  country: '',
  services: '',
  description: '',
  status: 'pending' as 'active' | 'inactive' | 'pending',
})

async function onSubmit(event: Event) {
  event.preventDefault()
  if (!form.vendorName || !form.email) {
    toast.error('Vendor name and email are required')
    return
  }

  isLoading.value = true
  try {
    await $fetch('/api/vendors', { method: 'POST', body: form })
    toast.success('Vendor created')
    emit('created')
  }
  catch (err: any) {
    toast.error(err?.data?.statusMessage || 'Failed to create vendor')
  }
  finally {
    isLoading.value = false
  }
}
</script>

<template>
  <form class="grid gap-4 px-4" @submit="onSubmit">
    <div class="grid gap-2">
      <Label for="vendorName">Vendor Name</Label>
      <Input id="vendorName" v-model="form.vendorName" :disabled="isLoading" />
    </div>
    <div class="grid gap-2">
      <Label for="email">Email</Label>
      <Input id="email" v-model="form.email" type="email" :disabled="isLoading" />
    </div>
    <div class="grid grid-cols-2 gap-4">
      <div class="grid gap-2">
        <Label for="phone">Phone</Label>
        <Input id="phone" v-model="form.phone" :disabled="isLoading" />
      </div>
      <div class="grid gap-2">
        <Label for="website">Website</Label>
        <Input id="website" v-model="form.website" :disabled="isLoading" />
      </div>
    </div>
    <div class="grid gap-2">
      <Label for="address">Address</Label>
      <Input id="address" v-model="form.address" :disabled="isLoading" />
    </div>
    <div class="grid grid-cols-2 gap-4">
      <div class="grid gap-2">
        <Label for="vendorType">Vendor Type</Label>
        <Input id="vendorType" v-model="form.vendorType" placeholder="e.g. Supplier" :disabled="isLoading" />
      </div>
      <div class="grid gap-2">
        <Label for="registrationNumber">Registration Number</Label>
        <Input id="registrationNumber" v-model="form.registrationNumber" :disabled="isLoading" />
      </div>
    </div>
    <div class="grid grid-cols-2 gap-4">
      <div class="grid gap-2">
        <Label for="foundedDate">Founded Date</Label>
        <Input id="foundedDate" v-model="form.foundedDate" type="date" :disabled="isLoading" />
      </div>
      <div class="grid gap-2">
        <Label for="country">Country</Label>
        <Input id="country" v-model="form.country" :disabled="isLoading" />
      </div>
    </div>
    <div class="grid gap-2">
      <Label for="services">Services</Label>
      <Input id="services" v-model="form.services" placeholder="Comma separated services" :disabled="isLoading" />
    </div>
    <div class="grid gap-2">
      <Label for="description">Description</Label>
      <Textarea id="description" v-model="form.description" :disabled="isLoading" />
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
      Create Vendor
    </Button>
  </form>
</template>
