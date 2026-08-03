<script setup lang="ts">
interface Payment {
  id: string
  invoiceId: string
  amount: string
  method: string
  reference?: string
  paidAt: string
  invoice?: { invoiceNumber: string, client?: { name: string } }
}

const { data: paymentsData, pending: isLoading } = await useFetch<Payment[]>('/api/payments')
const payments = computed(() => paymentsData.value ?? [])
</script>

<template>
  <FinanceLayout>
    <h3 class="text-lg font-medium">
      Payments
    </h3>
    <div class="rounded-md border">
      <Table>
        <TableHeader>
          <TableRow>
            <TableHead>Invoice #</TableHead>
            <TableHead>Client</TableHead>
            <TableHead>Amount</TableHead>
            <TableHead>Method</TableHead>
            <TableHead>Date</TableHead>
            <TableHead>Reference</TableHead>
          </TableRow>
        </TableHeader>
        <TableBody v-if="isLoading">
          <TableRow v-for="i in 5" :key="`skeleton-${i}`">
            <TableCell colspan="6">
              <Skeleton class="h-6 w-full" />
            </TableCell>
          </TableRow>
        </TableBody>
        <TableBody v-else>
          <TableRow v-for="payment in payments" :key="payment.id">
            <TableCell>
              <NuxtLink :to="`/finance/invoices/${payment.invoiceId}`" class="hover:underline">
                {{ payment.invoice?.invoiceNumber }}
              </NuxtLink>
            </TableCell>
            <TableCell>{{ payment.invoice?.client?.name || 'N/A' }}</TableCell>
            <TableCell>{{ Number(payment.amount).toFixed(2) }}</TableCell>
            <TableCell>{{ payment.method }}</TableCell>
            <TableCell>{{ payment.paidAt }}</TableCell>
            <TableCell>{{ payment.reference || '-' }}</TableCell>
          </TableRow>
          <TableRow v-if="payments.length === 0">
            <TableCell colspan="6" class="text-center text-muted-foreground">
              No payments recorded yet.
            </TableCell>
          </TableRow>
        </TableBody>
      </Table>
    </div>
  </FinanceLayout>
</template>
