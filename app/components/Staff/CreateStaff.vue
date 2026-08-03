<script setup lang="ts">
import { ref } from 'vue'
import { toast } from 'vue-sonner'

interface StaffForm {
  first_name: string
  last_name: string
  email: string
  phone_number: string
  role: string
  status: 'active' | 'inactive'
}

const form = ref<StaffForm>({
  first_name: '',
  last_name: '',
  email: '',
  phone_number: '',
  role: 'Staff',
  status: 'active',
})

const isSubmitting = ref(false)
const success = ref(false)

async function handleSubmit() {
  if (!form.value.first_name || !form.value.last_name || !form.value.email) {
    toast.error('Please fill in first name, last name and email')
    return
  }

  isSubmitting.value = true
  success.value = false
  try {
    // TODO: Replace with real API request
    await new Promise(resolve => setTimeout(resolve, 600))
    success.value = true
    toast.success('Staff created (mock)')
    // Reset basic fields
    form.value.first_name = ''
    form.value.last_name = ''
    form.value.email = ''
    form.value.phone_number = ''
    form.value.role = 'Staff'
    form.value.status = 'active'
  }
  finally {
    isSubmitting.value = false
  }
}
</script>

<template>
  <form class="space-y-4 pb-6" @submit.prevent="handleSubmit">
    <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
      <div class="space-y-2">
        <Label for="firstName">First name</Label>
        <Input id="firstName" v-model="form.first_name" placeholder="Jane" />
      </div>
      <div class="space-y-2">
        <Label for="lastName">Last name</Label>
        <Input id="lastName" v-model="form.last_name" placeholder="Doe" />
      </div>
    </div>

    <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
      <div class="space-y-2">
        <Label for="email">Email</Label>
        <Input id="email" v-model="form.email" type="email" placeholder="jane.doe@company.com" />
      </div>
      <div class="space-y-2">
        <Label for="phone">Phone</Label>
        <Input id="phone" v-model="form.phone_number" placeholder="+233 555 123 456" />
      </div>
    </div>

    <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
      <div class="space-y-2">
        <Label for="role">Role</Label>
        <Select v-model="form.role">
          <SelectTrigger id="role">
            <SelectValue placeholder="Select role" />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value="Staff">Staff</SelectItem>
            <SelectItem value="Manager">Manager</SelectItem>
            <SelectItem value="Admin">Admin</SelectItem>
          </SelectContent>
        </Select>
      </div>
      <div class="space-y-2">
        <Label for="status">Status</Label>
        <Select v-model="form.status">
          <SelectTrigger id="status">
            <SelectValue placeholder="Select status" />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value="active">Active</SelectItem>
            <SelectItem value="inactive">Inactive</SelectItem>
          </SelectContent>
        </Select>
      </div>
    </div>

    <div class="flex items-center gap-2 pt-2">
      <Button type="submit" :disabled="isSubmitting">
        <Icon v-if="isSubmitting" name="i-lucide-loader-2" class="mr-2 animate-spin" />
        Save Staff
      </Button>
      <p v-if="success" class="text-green-600 text-sm">Staff created (mock).</p>
    </div>
  </form>
</template>
