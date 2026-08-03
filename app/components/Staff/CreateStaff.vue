<script setup lang="ts">
import { ref } from 'vue'
import { toast } from 'vue-sonner'

interface StaffForm {
  firstName: string
  lastName: string
  email: string
  phoneNumber: string
  role: string
  status: 'active' | 'inactive'
  createLogin: boolean
}

const emit = defineEmits<{ created: [] }>()
const { user: authUser } = useAuth()
const isAdmin = computed(() => authUser.value?.role === 'admin')

const form = ref<StaffForm>({
  firstName: '',
  lastName: '',
  email: '',
  phoneNumber: '',
  role: 'Staff',
  status: 'active',
  createLogin: true,
})

const isSubmitting = ref(false)

async function handleSubmit() {
  if (!form.value.firstName || !form.value.lastName || !form.value.email) {
    toast.error('Please fill in first name, last name and email')
    return
  }

  isSubmitting.value = true
  try {
    const result = await $fetch<{ inviteEmailSent: boolean | null, inviteLink: string | null }>('/api/staff', {
      method: 'POST',
      body: { ...form.value, createLogin: isAdmin.value && form.value.createLogin },
    })

    if (result.inviteEmailSent === false) {
      toast.warning('Staff created, but the invite email could not be sent. Share this link with them directly:', {
        description: result.inviteLink ?? undefined,
        duration: 15000,
      })
    }
    else {
      toast.success(result.inviteEmailSent ? 'Staff created and invite sent' : 'Staff created')
    }
    emit('created')
  }
  catch (err: any) {
    toast.error(err?.data?.statusMessage || 'Failed to create staff')
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
        <Input id="firstName" v-model="form.firstName" placeholder="Jane" />
      </div>
      <div class="space-y-2">
        <Label for="lastName">Last name</Label>
        <Input id="lastName" v-model="form.lastName" placeholder="Doe" />
      </div>
    </div>

    <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
      <div class="space-y-2">
        <Label for="email">Email</Label>
        <Input id="email" v-model="form.email" type="email" placeholder="jane.doe@company.com" />
      </div>
      <div class="space-y-2">
        <Label for="phone">Phone</Label>
        <Input id="phone" v-model="form.phoneNumber" placeholder="+233 555 123 456" />
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
            <SelectItem value="Staff">
              Staff
            </SelectItem>
            <SelectItem value="Manager">
              Manager
            </SelectItem>
            <SelectItem value="Admin">
              Admin
            </SelectItem>
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
            <SelectItem value="active">
              Active
            </SelectItem>
            <SelectItem value="inactive">
              Inactive
            </SelectItem>
          </SelectContent>
        </Select>
      </div>
    </div>

    <div v-if="isAdmin" class="flex items-center gap-2">
      <Checkbox id="createLogin" v-model="form.createLogin" />
      <Label for="createLogin">Give this staff member login access (emails them an invite to set their password)</Label>
    </div>

    <div class="flex items-center gap-2 pt-2">
      <Button type="submit" :disabled="isSubmitting">
        <Icon v-if="isSubmitting" name="i-lucide-loader-2" class="mr-2 animate-spin" />
        Save Staff
      </Button>
    </div>
  </form>
</template>
