<script setup lang="ts">
definePageMeta({
  layout: 'blank',
})

interface LineItem {
  id: string
  description: string
  quantity: string
  unitPrice: string
  amount: string
}

interface TaxLine {
  id: string
  name: string
  rate: string
  amount: string
  compound: boolean
}

interface Payment {
  id: string
  amount: string
  method: string
  reference?: string
  paidAt: string
}

interface InvoiceDetail {
  id: string
  invoiceNumber: string
  receiptNumber?: string | null
  issueDate: string
  dueDate: string
  currency: string
  subtotal: string
  taxExempt: boolean
  taxAmount: string
  discount: string
  total: string
  amountPaid: number
  balanceDue: number
  client?: { name: string, email: string, phone?: string, address?: string }
  lineItems: LineItem[]
  taxes: TaxLine[]
  payments: Payment[]
}

interface CompanySettings {
  companyName: string
  tin?: string | null
  vatNumber?: string | null
  address?: string | null
  phone?: string | null
  email?: string | null
}

const route = useRoute()
const { data: invoice } = await useFetch<InvoiceDetail>(`/api/invoices/${route.params.id}`)
const { data: taxSettings } = await useFetch<{ settings: CompanySettings | null }>('/api/tax-settings')

function formatMoney(value: number | string) {
  return new Intl.NumberFormat('en-GH', { minimumFractionDigits: 2, maximumFractionDigits: 2 }).format(Number(value))
}

function print() {
  window.print()
}
</script>

<template>
  <div class="receipt-page mx-auto max-w-2xl p-8 print:p-0">
    <div class="mb-6 flex justify-end gap-2 print:hidden">
      <Button variant="outline" @click="print">
        <Icon name="i-lucide-printer" />
        Print / Save as PDF
      </Button>
    </div>

    <div v-if="invoice" class="rounded-md border bg-background p-8 print:border-none print:p-0">
      <div class="mb-6 flex items-start justify-between border-b pb-6">
        <div>
          <h1 class="text-xl font-bold">
            {{ taxSettings?.settings?.companyName || 'Ingenicx' }}
          </h1>
          <div class="mt-1 text-sm text-muted-foreground space-y-0.5">
            <div v-if="taxSettings?.settings?.address">
              {{ taxSettings.settings.address }}
            </div>
            <div v-if="taxSettings?.settings?.phone">
              Tel: {{ taxSettings.settings.phone }}
            </div>
            <div v-if="taxSettings?.settings?.email">
              {{ taxSettings.settings.email }}
            </div>
            <div v-if="taxSettings?.settings?.tin">
              TIN: {{ taxSettings.settings.tin }}
            </div>
            <div v-if="taxSettings?.settings?.vatNumber">
              VAT Reg. No: {{ taxSettings.settings.vatNumber }}
            </div>
          </div>
        </div>
        <div class="text-right">
          <h2 class="text-2xl font-bold tracking-tight">
            RECEIPT
          </h2>
          <div class="mt-1 text-sm">
            <div><span class="text-muted-foreground">Receipt No:</span> {{ invoice.receiptNumber || 'N/A' }}</div>
            <div><span class="text-muted-foreground">Invoice No:</span> {{ invoice.invoiceNumber }}</div>
            <div><span class="text-muted-foreground">Date:</span> {{ invoice.issueDate }}</div>
          </div>
        </div>
      </div>

      <div class="mb-6">
        <div class="text-xs font-semibold uppercase text-muted-foreground">
          Billed To
        </div>
        <div class="font-medium">
          {{ invoice.client?.name }}
        </div>
        <div class="text-sm text-muted-foreground">
          {{ invoice.client?.email }}
        </div>
      </div>

      <table class="w-full text-sm">
        <thead>
          <tr class="border-b text-left">
            <th class="py-2">
              Description
            </th>
            <th class="py-2">
              Qty
            </th>
            <th class="py-2">
              Unit Price
            </th>
            <th class="py-2 text-right">
              Amount
            </th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="item in invoice.lineItems" :key="item.id" class="border-b">
            <td class="py-2">
              {{ item.description }}
            </td>
            <td class="py-2">
              {{ item.quantity }}
            </td>
            <td class="py-2">
              {{ formatMoney(item.unitPrice) }}
            </td>
            <td class="py-2 text-right">
              {{ formatMoney(item.amount) }}
            </td>
          </tr>
        </tbody>
      </table>

      <div class="mt-4 flex justify-end">
        <div class="w-64 space-y-1 text-sm">
          <div class="flex justify-between">
            <span class="text-muted-foreground">Subtotal</span>
            <span>{{ formatMoney(invoice.subtotal) }}</span>
          </div>
          <div class="flex justify-between">
            <span class="text-muted-foreground">Discount</span>
            <span>-{{ formatMoney(invoice.discount) }}</span>
          </div>
          <div v-if="invoice.taxExempt" class="flex justify-between">
            <span class="text-muted-foreground">Tax</span>
            <span>Exempt</span>
          </div>
          <div v-for="tax in invoice.taxes" v-else :key="tax.id" class="flex justify-between">
            <span class="text-muted-foreground">{{ tax.name }} ({{ Number(tax.rate) }}%)</span>
            <span>{{ formatMoney(tax.amount) }}</span>
          </div>
          <div class="flex justify-between border-t pt-1 font-semibold">
            <span>Total</span>
            <span>{{ invoice.currency }} {{ formatMoney(invoice.total) }}</span>
          </div>
          <div class="flex justify-between text-green-700">
            <span>Amount Paid</span>
            <span>{{ invoice.currency }} {{ formatMoney(invoice.amountPaid) }}</span>
          </div>
          <div class="flex justify-between font-semibold">
            <span>Balance Due</span>
            <span>{{ invoice.currency }} {{ formatMoney(invoice.balanceDue) }}</span>
          </div>
        </div>
      </div>

      <div class="mt-8">
        <div class="text-xs font-semibold uppercase text-muted-foreground mb-2">
          Payments Received
        </div>
        <table class="w-full text-sm">
          <thead>
            <tr class="border-b text-left">
              <th class="py-2">
                Date
              </th>
              <th class="py-2">
                Method
              </th>
              <th class="py-2">
                Reference
              </th>
              <th class="py-2 text-right">
                Amount
              </th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="payment in invoice.payments" :key="payment.id" class="border-b">
              <td class="py-2">
                {{ payment.paidAt }}
              </td>
              <td class="py-2">
                {{ payment.method.replace('_', ' ') }}
              </td>
              <td class="py-2">
                {{ payment.reference || '-' }}
              </td>
              <td class="py-2 text-right">
                {{ formatMoney(payment.amount) }}
              </td>
            </tr>
          </tbody>
        </table>
      </div>

      <div class="mt-10 text-center text-xs text-muted-foreground">
        This receipt was generated electronically and is valid without a signature.
      </div>
    </div>
  </div>
</template>

<style>
@media print {
  body {
    background: white;
  }
}
</style>
