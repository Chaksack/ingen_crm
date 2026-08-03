<script setup lang="ts">
import { toast } from 'vue-sonner'

interface LineItem {
  id: string
  description: string
  quantity: string
  unitPrice: string
  amount: string
}

interface Payment {
  id: string
  amount: string
  method: string
  reference?: string
  paidAt: string
}

interface TaxLine {
  id: string
  name: string
  rate: string
  amount: string
  compound: boolean
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
  notes?: string
  terms?: string
  effectiveStatus: string
  balanceDue: number
  amountPaid: number
  client?: { name: string, email: string }
  lineItems: LineItem[]
  taxes: TaxLine[]
  payments: Payment[]
}

const route = useRoute()
const { data: invoice, refresh } = await useFetch<InvoiceDetail>(`/api/invoices/${route.params.id}`)

const isPaymentDialogOpen = ref(false)
const isSubmittingPayment = ref(false)
const paymentForm = reactive({
  amount: 0,
  method: 'bank_transfer',
  reference: '',
  paidAt: new Date().toISOString().slice(0, 10),
})

function getBadgeVariant(status: string) {
  switch (status) {
    case 'paid':
      return 'default'
    case 'overdue':
      return 'destructive'
    case 'void':
      return 'secondary'
    default:
      return 'secondary'
  }
}

async function recordPayment(event: Event) {
  event.preventDefault()
  if (paymentForm.amount <= 0) {
    toast.error('Enter a valid payment amount')
    return
  }

  isSubmittingPayment.value = true
  try {
    await $fetch('/api/payments', {
      method: 'POST',
      body: { ...paymentForm, invoiceId: route.params.id },
    })
    toast.success('Payment recorded')
    isPaymentDialogOpen.value = false
    paymentForm.amount = 0
    paymentForm.reference = ''
    await refresh()
  }
  catch (err: any) {
    toast.error(err?.data?.statusMessage || 'Failed to record payment')
  }
  finally {
    isSubmittingPayment.value = false
  }
}

async function markSent() {
  await $fetch(`/api/invoices/${route.params.id}`, { method: 'PATCH', body: { status: 'sent' } })
  await refresh()
  toast.success('Invoice marked as sent')
}
</script>

<template>
  <div class="w-full flex flex-col gap-4">
    <div class="flex items-center gap-4 flex-wrap">
      <NuxtLink to="/finance/invoices">
        <Button variant="outline" size="icon" class="h-7 w-7">
          <Icon name="i-lucide-chevron-left" />
        </Button>
      </NuxtLink>
      <h2 class="text-2xl font-bold tracking-tight">
        {{ invoice?.invoiceNumber }}
      </h2>
      <Badge v-if="invoice" :variant="getBadgeVariant(invoice.effectiveStatus)">
        {{ invoice.effectiveStatus }}
      </Badge>
      <div class="ml-auto flex gap-2">
        <Button v-if="invoice?.effectiveStatus === 'draft'" variant="outline" @click="markSent">
          Mark as Sent
        </Button>
        <NuxtLink v-if="invoice && invoice.amountPaid > 0" :to="`/finance/invoices/${invoice.id}/receipt`" target="_blank">
          <Button variant="outline">
            <Icon name="i-lucide-receipt" />
            View Receipt
          </Button>
        </NuxtLink>
        <Dialog v-model:open="isPaymentDialogOpen">
          <DialogTrigger as-child>
            <Button>
              Record Payment
            </Button>
          </DialogTrigger>
          <DialogContent class="sm:max-w-md">
            <DialogHeader>
              <DialogTitle>Record Payment</DialogTitle>
              <DialogDescription>
                Record a payment against this invoice.
              </DialogDescription>
            </DialogHeader>
            <form class="grid gap-4" @submit="recordPayment">
              <div class="grid gap-2">
                <Label for="amount">Amount</Label>
                <Input id="amount" v-model.number="paymentForm.amount" type="number" min="0" step="0.01" />
              </div>
              <div class="grid gap-2">
                <Label for="method">Method</Label>
                <Select v-model="paymentForm.method">
                  <SelectTrigger id="method" class="w-full">
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
                <Label for="paidAt">Date</Label>
                <Input id="paidAt" v-model="paymentForm.paidAt" type="date" />
              </div>
              <div class="grid gap-2">
                <Label for="reference">Reference</Label>
                <Input id="reference" v-model="paymentForm.reference" placeholder="Optional" />
              </div>
              <DialogFooter>
                <Button type="submit" :disabled="isSubmittingPayment">
                  Save Payment
                </Button>
              </DialogFooter>
            </form>
          </DialogContent>
        </Dialog>
      </div>
    </div>

    <div v-if="invoice" class="grid gap-4 md:grid-cols-3">
      <Card class="md:col-span-2">
        <CardHeader>
          <CardTitle>Line Items</CardTitle>
          <CardDescription>Billed to {{ invoice.client?.name }} ({{ invoice.client?.email }})</CardDescription>
        </CardHeader>
        <CardContent>
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>Description</TableHead>
                <TableHead>Qty</TableHead>
                <TableHead>Unit Price</TableHead>
                <TableHead class="text-right">
                  Amount
                </TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              <TableRow v-for="item in invoice.lineItems" :key="item.id">
                <TableCell>{{ item.description }}</TableCell>
                <TableCell>{{ item.quantity }}</TableCell>
                <TableCell>{{ Number(item.unitPrice).toFixed(2) }}</TableCell>
                <TableCell class="text-right">
                  {{ Number(item.amount).toFixed(2) }}
                </TableCell>
              </TableRow>
            </TableBody>
          </Table>
        </CardContent>
        <CardFooter class="flex flex-col items-end gap-1 text-sm">
          <div class="flex justify-between w-48">
            <span class="text-muted-foreground">Subtotal</span>
            <span>{{ Number(invoice.subtotal).toFixed(2) }}</span>
          </div>
          <div class="flex justify-between w-48">
            <span class="text-muted-foreground">Discount</span>
            <span>-{{ Number(invoice.discount).toFixed(2) }}</span>
          </div>
          <div v-if="invoice.taxExempt" class="flex justify-between w-48">
            <span class="text-muted-foreground">Tax</span>
            <span>Exempt</span>
          </div>
          <div v-for="tax in invoice.taxes" v-else :key="tax.id" class="flex justify-between w-48">
            <span class="text-muted-foreground">{{ tax.name }} ({{ Number(tax.rate) }}%)</span>
            <span>{{ Number(tax.amount).toFixed(2) }}</span>
          </div>
          <div class="flex justify-between w-48 font-semibold border-t pt-1">
            <span>Total</span>
            <span>{{ Number(invoice.total).toFixed(2) }}</span>
          </div>
          <div class="flex justify-between w-48 text-green-600">
            <span>Paid</span>
            <span>{{ invoice.amountPaid.toFixed(2) }}</span>
          </div>
          <div class="flex justify-between w-48 font-semibold">
            <span>Balance Due</span>
            <span>{{ invoice.balanceDue.toFixed(2) }}</span>
          </div>
        </CardFooter>
      </Card>

      <Card>
        <CardHeader>
          <CardTitle>Payments</CardTitle>
        </CardHeader>
        <CardContent class="space-y-3">
          <div v-if="invoice.payments.length === 0" class="text-sm text-muted-foreground">
            No payments recorded yet.
          </div>
          <div v-for="payment in invoice.payments" :key="payment.id" class="flex justify-between text-sm border-b pb-2">
            <div>
              <div class="font-medium">
                {{ Number(payment.amount).toFixed(2) }}
              </div>
              <div class="text-muted-foreground">
                {{ payment.method }} &middot; {{ payment.paidAt }}
              </div>
            </div>
            <div v-if="payment.reference" class="text-muted-foreground">
              {{ payment.reference }}
            </div>
          </div>
        </CardContent>
      </Card>

      <Card v-if="invoice.notes || invoice.terms" class="md:col-span-3">
        <CardContent class="grid md:grid-cols-2 gap-4 pt-6">
          <div v-if="invoice.notes">
            <div class="text-sm text-muted-foreground mb-1">
              Notes
            </div>
            <div class="text-sm">
              {{ invoice.notes }}
            </div>
          </div>
          <div v-if="invoice.terms">
            <div class="text-sm text-muted-foreground mb-1">
              Terms
            </div>
            <div class="text-sm">
              {{ invoice.terms }}
            </div>
          </div>
        </CardContent>
      </Card>
    </div>
  </div>
</template>
