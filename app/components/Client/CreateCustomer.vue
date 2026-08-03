<script setup lang="ts">
import { Loader2, Plus, Trash2 } from 'lucide-vue-next'
import { toast } from 'vue-sonner'
import { Button } from '@/components/ui/button'
import { Checkbox } from '@/components/ui/checkbox'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Textarea } from '@/components/ui/textarea'

const emit = defineEmits<{ created: [] }>()

const isLoading = ref(false)
const form = reactive({
  name: '',
  contactPerson: '',
  email: '',
  phone: '',
  address: '',
  notes: '',
  dob: '',
  nationality: '',
  idType: '',
  idNumber: '',
  monthlyIncome: '',
  kyc: false,
})

const bankAccounts = ref<{ bankName: string, accountNumber: string, accountHolder: string }[]>([])
const momoAccounts = ref<{ networkName: string, momoNumber: string, accountHolder: string }[]>([])

function addBankAccount() {
  bankAccounts.value.push({ bankName: '', accountNumber: '', accountHolder: '' })
}

function addMomoAccount() {
  momoAccounts.value.push({ networkName: '', momoNumber: '', accountHolder: '' })
}

async function onSubmit(event: Event) {
  event.preventDefault()
  if (!form.name || !form.email) {
    toast.error('Client name and email are required')
    return
  }

  isLoading.value = true
  try {
    await $fetch('/api/clients', {
      method: 'POST',
      body: {
        ...form,
        monthlyIncome: form.monthlyIncome ? Number(form.monthlyIncome) : undefined,
        bankAccounts: bankAccounts.value.filter(a => a.bankName && a.accountNumber),
        momoAccounts: momoAccounts.value.filter(a => a.networkName && a.momoNumber),
      },
    })
    toast.success('Client created')
    emit('created')
  }
  catch (err: any) {
    toast.error(err?.data?.statusMessage || 'Failed to create client')
  }
  finally {
    isLoading.value = false
  }
}
</script>

<template>
  <form class="grid gap-4 px-4" @submit="onSubmit">
    <div class="grid grid-cols-2 gap-4">
      <div class="grid gap-2">
        <Label for="name">Client Name</Label>
        <Input id="name" v-model="form.name" :disabled="isLoading" />
      </div>
      <div class="grid gap-2">
        <Label for="contactPerson">Contact Person</Label>
        <Input id="contactPerson" v-model="form.contactPerson" :disabled="isLoading" />
      </div>
    </div>
    <div class="grid grid-cols-2 gap-4">
      <div class="grid gap-2">
        <Label for="email">Email</Label>
        <Input id="email" v-model="form.email" type="email" :disabled="isLoading" />
      </div>
      <div class="grid gap-2">
        <Label for="phone">Phone</Label>
        <Input id="phone" v-model="form.phone" :disabled="isLoading" />
      </div>
    </div>
    <div class="grid gap-2">
      <Label for="address">Address</Label>
      <Input id="address" v-model="form.address" :disabled="isLoading" />
    </div>

    <div class="grid grid-cols-3 gap-4">
      <div class="grid gap-2">
        <Label for="dob">Date of Birth</Label>
        <Input id="dob" v-model="form.dob" type="date" :disabled="isLoading" />
      </div>
      <div class="grid gap-2">
        <Label for="nationality">Nationality</Label>
        <Input id="nationality" v-model="form.nationality" :disabled="isLoading" />
      </div>
      <div class="grid gap-2">
        <Label for="monthlyIncome">Monthly Income</Label>
        <Input id="monthlyIncome" v-model="form.monthlyIncome" type="number" min="0" step="0.01" :disabled="isLoading" />
      </div>
    </div>
    <div class="grid grid-cols-2 gap-4">
      <div class="grid gap-2">
        <Label for="idType">ID Type</Label>
        <Input id="idType" v-model="form.idType" placeholder="e.g. Ghana Card" :disabled="isLoading" />
      </div>
      <div class="grid gap-2">
        <Label for="idNumber">ID Number</Label>
        <Input id="idNumber" v-model="form.idNumber" :disabled="isLoading" />
      </div>
    </div>
    <div class="flex items-center gap-2">
      <Checkbox id="kyc" v-model="form.kyc" />
      <Label for="kyc">KYC Verified</Label>
    </div>
    <div class="grid gap-2">
      <Label for="notes">Notes</Label>
      <Textarea id="notes" v-model="form.notes" :disabled="isLoading" />
    </div>

    <div class="grid gap-2">
      <div class="flex items-center justify-between">
        <Label>Bank Accounts</Label>
        <Button type="button" size="sm" variant="outline" @click="addBankAccount">
          <Plus class="h-4 w-4" /> Add
        </Button>
      </div>
      <div v-for="(account, index) in bankAccounts" :key="index" class="grid grid-cols-[1fr_1fr_1fr_auto] gap-2 items-end">
        <Input v-model="account.bankName" placeholder="Bank name" />
        <Input v-model="account.accountNumber" placeholder="Account number" />
        <Input v-model="account.accountHolder" placeholder="Account holder" />
        <Button type="button" size="icon" variant="ghost" @click="bankAccounts.splice(index, 1)">
          <Trash2 class="h-4 w-4" />
        </Button>
      </div>
    </div>

    <div class="grid gap-2">
      <div class="flex items-center justify-between">
        <Label>Mobile Money Accounts</Label>
        <Button type="button" size="sm" variant="outline" @click="addMomoAccount">
          <Plus class="h-4 w-4" /> Add
        </Button>
      </div>
      <div v-for="(account, index) in momoAccounts" :key="index" class="grid grid-cols-[1fr_1fr_1fr_auto] gap-2 items-end">
        <Input v-model="account.networkName" placeholder="Network" />
        <Input v-model="account.momoNumber" placeholder="Mobile number" />
        <Input v-model="account.accountHolder" placeholder="Account holder" />
        <Button type="button" size="icon" variant="ghost" @click="momoAccounts.splice(index, 1)">
          <Trash2 class="h-4 w-4" />
        </Button>
      </div>
    </div>

    <Button type="submit" :disabled="isLoading" class="mt-2">
      <Loader2 v-if="isLoading" class="mr-2 h-4 w-4 animate-spin" />
      Create Client
    </Button>
  </form>
</template>
