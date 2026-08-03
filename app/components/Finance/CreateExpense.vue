<script setup lang="ts">
import { Loader2 } from 'lucide-vue-next'
import { toast } from 'vue-sonner'

interface VendorOption {
  id: string
  vendorName: string
}

const emit = defineEmits<{ created: [] }>()

const { data: vendors } = await useFetch<VendorOption[]>('/api/vendors')

const isLoading = ref(false)
const receiptFile = ref<File | null>(null)
const form = reactive({
  category: '',
  vendorId: '',
  amount: 0,
  currency: 'GHS',
  expenseDate: new Date().toISOString().slice(0, 10),
  description: '',
  paymentMethod: 'cash',
})

function onFileChange(event: Event) {
  const target = event.target as HTMLInputElement
  receiptFile.value = target.files?.[0] ?? null
}

async function onSubmit(event: Event) {
  event.preventDefault()
  if (!form.category || form.amount <= 0) {
    toast.error('Category and a valid amount are required')
    return
  }

  isLoading.value = true
  try {
    let receiptUrl: string | undefined
    if (receiptFile.value) {
      const body = new FormData()
      body.append('file', receiptFile.value)
      const uploaded = await $fetch<{ url: string }>('/api/expenses/upload', { method: 'POST', body })
      receiptUrl = uploaded.url
    }

    await $fetch('/api/expenses', {
      method: 'POST',
      body: {
        ...form,
        vendorId: form.vendorId || undefined,
        receiptUrl,
      },
    })
    toast.success('Expense recorded')
    emit('created')
  }
  catch (err: any) {
    toast.error(err?.data?.statusMessage || 'Failed to record expense')
  }
  finally {
    isLoading.value = false
  }
}
</script>

<template>
  <form class="grid gap-4 px-4" @submit="onSubmit">
    <div class="grid gap-2">
      <Label for="category">Category</Label>
      <Input id="category" v-model="form.category" placeholder="e.g. Office Supplies" :disabled="isLoading" />
    </div>
    <div class="grid gap-2">
      <Label for="vendorId">Vendor (optional)</Label>
      <Select v-model="form.vendorId">
        <SelectTrigger id="vendorId" class="w-full">
          <SelectValue placeholder="Select a vendor" />
        </SelectTrigger>
        <SelectContent>
          <SelectItem v-for="vendor in vendors" :key="vendor.id" :value="vendor.id">
            {{ vendor.vendorName }}
          </SelectItem>
        </SelectContent>
      </Select>
    </div>
    <div class="grid grid-cols-2 gap-4">
      <div class="grid gap-2">
        <Label for="amount">Amount</Label>
        <Input id="amount" v-model.number="form.amount" type="number" min="0" step="0.01" :disabled="isLoading" />
      </div>
      <div class="grid gap-2">
        <Label for="expenseDate">Date</Label>
        <Input id="expenseDate" v-model="form.expenseDate" type="date" :disabled="isLoading" />
      </div>
    </div>
    <div class="grid gap-2">
      <Label for="paymentMethod">Payment Method</Label>
      <Select v-model="form.paymentMethod">
        <SelectTrigger id="paymentMethod" class="w-full">
          <SelectValue />
        </SelectTrigger>
        <SelectContent>
          <SelectItem value="cash">
            Cash
          </SelectItem>
          <SelectItem value="bank_transfer">
            Bank Transfer
          </SelectItem>
          <SelectItem value="mobile_money">
            Mobile Money
          </SelectItem>
          <SelectItem value="card">
            Card
          </SelectItem>
          <SelectItem value="cheque">
            Cheque
          </SelectItem>
        </SelectContent>
      </Select>
    </div>
    <div class="grid gap-2">
      <Label for="description">Description</Label>
      <Textarea id="description" v-model="form.description" :disabled="isLoading" />
    </div>
    <div class="grid gap-2">
      <Label for="receipt">Receipt (optional)</Label>
      <Input id="receipt" type="file" accept="image/*,application/pdf" :disabled="isLoading" @change="onFileChange" />
    </div>
    <Button type="submit" :disabled="isLoading" class="mt-2">
      <Loader2 v-if="isLoading" class="mr-2 h-4 w-4 animate-spin" />
      Record Expense
    </Button>
  </form>
</template>
