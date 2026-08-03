<script setup lang="ts">
import { Loader2, Plus, Trash2 } from 'lucide-vue-next'
import { toast } from 'vue-sonner'

interface ClientOption {
  id: string
  name: string
}

const emit = defineEmits<{ created: [] }>()

const { data: clients } = await useFetch<ClientOption[]>('/api/clients')
const { activeRates } = useTaxPreview()

const documentType = ref<'invoice' | 'quotation'>('invoice')
const isLoading = ref(false)
const form = reactive({
  clientId: '',
  issueDate: new Date().toISOString().slice(0, 10),
  secondDate: new Date(Date.now() + 14 * 24 * 60 * 60 * 1000).toISOString().slice(0, 10),
  currency: 'GHS',
  taxExempt: false,
  discount: 0,
  notes: '',
  terms: '',
})

const secondDateLabel = computed(() => documentType.value === 'invoice' ? 'Due Date' : 'Expiry Date')

const lineItems = ref([{ description: '', quantity: 1, unitPrice: 0 }])

function addLineItem() {
  lineItems.value.push({ description: '', quantity: 1, unitPrice: 0 })
}

const subtotal = computed(() => lineItems.value.reduce((sum, i) => sum + (i.quantity * i.unitPrice), 0))
const taxableBase = computed(() => Math.max(subtotal.value - form.discount, 0))
const taxPreview = computed(() => form.taxExempt
  ? { lines: [], taxAmount: 0 }
  : computeTaxPreview(taxableBase.value, activeRates.value))
const total = computed(() => taxableBase.value + taxPreview.value.taxAmount)

async function onSubmit(event: Event) {
  event.preventDefault()
  if (!form.clientId) {
    toast.error('Select a client')
    return
  }
  const items = lineItems.value.filter(i => i.description && i.quantity > 0)
  if (!items.length) {
    toast.error('Add at least one line item')
    return
  }

  const isInvoice = documentType.value === 'invoice'
  const { secondDate, ...rest } = form
  const body = isInvoice
    ? { ...rest, dueDate: secondDate, lineItems: items }
    : { ...rest, expiryDate: secondDate, lineItems: items }

  isLoading.value = true
  try {
    await $fetch(isInvoice ? '/api/invoices' : '/api/quotations', {
      method: 'POST',
      body,
    })
    toast.success(isInvoice ? 'Invoice created' : 'Quotation created')
    emit('created')
  }
  catch (err: any) {
    toast.error(err?.data?.statusMessage || `Failed to create ${documentType.value}`)
  }
  finally {
    isLoading.value = false
  }
}
</script>

<template>
  <form class="grid gap-4 px-4" @submit="onSubmit">
    <div class="grid gap-2">
      <Label for="documentType">Document Type</Label>
      <Select v-model="documentType">
        <SelectTrigger id="documentType" class="w-full">
          <SelectValue placeholder="Select document type" />
        </SelectTrigger>
        <SelectContent>
          <SelectItem value="invoice">
            Invoice
          </SelectItem>
          <SelectItem value="quotation">
            Quotation
          </SelectItem>
        </SelectContent>
      </Select>
    </div>

    <div class="grid gap-2">
      <Label for="clientId">Client</Label>
      <Select v-model="form.clientId">
        <SelectTrigger id="clientId" class="w-full">
          <SelectValue placeholder="Select a client" />
        </SelectTrigger>
        <SelectContent>
          <SelectItem v-for="client in clients" :key="client.id" :value="client.id">
            {{ client.name }}
          </SelectItem>
        </SelectContent>
      </Select>
    </div>

    <div class="grid grid-cols-2 gap-4">
      <div class="grid gap-2">
        <Label for="issueDate">Issue Date</Label>
        <Input id="issueDate" v-model="form.issueDate" type="date" />
      </div>
      <div class="grid gap-2">
        <Label for="secondDate">{{ secondDateLabel }}</Label>
        <Input id="secondDate" v-model="form.secondDate" type="date" />
      </div>
    </div>

    <div class="grid gap-2">
      <div class="flex items-center justify-between">
        <Label>Line Items</Label>
        <Button type="button" size="sm" variant="outline" @click="addLineItem">
          <Plus class="h-4 w-4" /> Add Item
        </Button>
      </div>
      <div v-for="(item, index) in lineItems" :key="index" class="grid grid-cols-[2fr_1fr_1fr_auto] gap-2 items-end">
        <Input v-model="item.description" placeholder="Description" />
        <Input v-model.number="item.quantity" type="number" min="0" step="0.01" placeholder="Qty" />
        <Input v-model.number="item.unitPrice" type="number" min="0" step="0.01" placeholder="Unit price" />
        <Button type="button" size="icon" variant="ghost" :disabled="lineItems.length === 1" @click="lineItems.splice(index, 1)">
          <Trash2 class="h-4 w-4" />
        </Button>
      </div>
    </div>

    <div class="grid grid-cols-2 gap-4 items-end">
      <div class="grid gap-2">
        <Label for="discount">Discount</Label>
        <Input id="discount" v-model.number="form.discount" type="number" min="0" step="0.01" />
      </div>
      <div class="flex items-center gap-2 pb-2">
        <Checkbox id="taxExempt" v-model="form.taxExempt" />
        <Label for="taxExempt">Tax Exempt</Label>
      </div>
    </div>

    <div class="rounded-md border p-4 space-y-1 text-sm">
      <div class="flex justify-between">
        <span class="text-muted-foreground">Subtotal</span>
        <span>{{ subtotal.toFixed(2) }}</span>
      </div>
      <div class="flex justify-between">
        <span class="text-muted-foreground">Discount</span>
        <span>-{{ form.discount.toFixed(2) }}</span>
      </div>
      <div v-if="!form.taxExempt && taxPreview.lines.length === 0" class="text-xs text-muted-foreground">
        No active tax rates configured. <NuxtLink to="/finance/tax-settings" class="underline">
          Set up tax rates
        </NuxtLink>
      </div>
      <div v-for="line in taxPreview.lines" :key="line.name" class="flex justify-between">
        <span class="text-muted-foreground">{{ line.name }} ({{ line.rate }}%)</span>
        <span>{{ line.amount.toFixed(2) }}</span>
      </div>
      <div class="flex justify-between font-semibold pt-1 border-t">
        <span>Total</span>
        <span>{{ total.toFixed(2) }}</span>
      </div>
    </div>

    <div class="grid gap-2">
      <Label for="notes">Notes</Label>
      <Textarea id="notes" v-model="form.notes" />
    </div>
    <div class="grid gap-2">
      <Label for="terms">Terms</Label>
      <Textarea id="terms" v-model="form.terms" />
    </div>

    <Button type="submit" :disabled="isLoading" class="mt-2">
      <Loader2 v-if="isLoading" class="mr-2 h-4 w-4 animate-spin" />
      {{ documentType === 'invoice' ? 'Create Invoice' : 'Create Quotation' }}
    </Button>
  </form>
</template>
