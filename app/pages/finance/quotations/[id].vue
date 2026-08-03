<script setup lang="ts">
import { toast } from 'vue-sonner'

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

interface QuotationDetail {
  id: string
  quoteNumber: string
  issueDate: string
  expiryDate: string
  subtotal: string
  taxExempt: boolean
  taxAmount: string
  discount: string
  total: string
  notes?: string
  terms?: string
  status: string
  convertedInvoiceId?: string | null
  client?: { name: string, email: string }
  lineItems: LineItem[]
  taxes: TaxLine[]
}

const route = useRoute()
const { data: quotation } = await useFetch<QuotationDetail>(`/api/quotations/${route.params.id}`)

const isConverting = ref(false)

function getBadgeVariant(status: string) {
  switch (status) {
    case 'accepted':
      return 'default'
    case 'rejected':
    case 'expired':
      return 'destructive'
    default:
      return 'secondary'
  }
}

async function convertToInvoice() {
  isConverting.value = true
  try {
    const invoice = await $fetch<{ id: string }>(`/api/quotations/${route.params.id}/convert`, {
      method: 'POST',
      body: {
        issueDate: new Date().toISOString().slice(0, 10),
        dueDate: new Date(Date.now() + 14 * 24 * 60 * 60 * 1000).toISOString().slice(0, 10),
      },
    })
    toast.success('Converted to invoice')
    await navigateTo(`/finance/invoices/${invoice.id}`)
  }
  catch (err: any) {
    toast.error(err?.data?.statusMessage || 'Failed to convert quotation')
  }
  finally {
    isConverting.value = false
  }
}
</script>

<template>
  <div class="w-full flex flex-col gap-4">
    <div class="flex items-center gap-4 flex-wrap">
      <NuxtLink to="/finance/quotations">
        <Button variant="outline" size="icon" class="h-7 w-7">
          <Icon name="i-lucide-chevron-left" />
        </Button>
      </NuxtLink>
      <h2 class="text-2xl font-bold tracking-tight">
        {{ quotation?.quoteNumber }}
      </h2>
      <Badge v-if="quotation" :variant="getBadgeVariant(quotation.status)">
        {{ quotation.status }}
      </Badge>
      <div class="ml-auto">
        <Button
          v-if="quotation && !quotation.convertedInvoiceId"
          :disabled="isConverting"
          @click="convertToInvoice"
        >
          Convert to Invoice
        </Button>
        <NuxtLink v-else-if="quotation?.convertedInvoiceId" :to="`/finance/invoices/${quotation.convertedInvoiceId}`">
          <Button variant="outline">
            View Invoice
          </Button>
        </NuxtLink>
      </div>
    </div>

    <div v-if="quotation" class="grid gap-4">
      <Card>
        <CardHeader>
          <CardTitle>Line Items</CardTitle>
          <CardDescription>For {{ quotation.client?.name }} ({{ quotation.client?.email }})</CardDescription>
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
              <TableRow v-for="item in quotation.lineItems" :key="item.id">
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
            <span>{{ Number(quotation.subtotal).toFixed(2) }}</span>
          </div>
          <div class="flex justify-between w-48">
            <span class="text-muted-foreground">Discount</span>
            <span>-{{ Number(quotation.discount).toFixed(2) }}</span>
          </div>
          <div v-if="quotation.taxExempt" class="flex justify-between w-48">
            <span class="text-muted-foreground">Tax</span>
            <span>Exempt</span>
          </div>
          <div v-for="tax in quotation.taxes" v-else :key="tax.id" class="flex justify-between w-48">
            <span class="text-muted-foreground">{{ tax.name }} ({{ Number(tax.rate) }}%)</span>
            <span>{{ Number(tax.amount).toFixed(2) }}</span>
          </div>
          <div class="flex justify-between w-48 font-semibold border-t pt-1">
            <span>Total</span>
            <span>{{ Number(quotation.total).toFixed(2) }}</span>
          </div>
        </CardFooter>
      </Card>

      <Card v-if="quotation.notes || quotation.terms">
        <CardContent class="grid md:grid-cols-2 gap-4 pt-6">
          <div v-if="quotation.notes">
            <div class="text-sm text-muted-foreground mb-1">
              Notes
            </div>
            <div class="text-sm">
              {{ quotation.notes }}
            </div>
          </div>
          <div v-if="quotation.terms">
            <div class="text-sm text-muted-foreground mb-1">
              Terms
            </div>
            <div class="text-sm">
              {{ quotation.terms }}
            </div>
          </div>
        </CardContent>
      </Card>
    </div>
  </div>
</template>
