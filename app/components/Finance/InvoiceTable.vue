<script setup lang="ts">
interface Invoice {
  id: string
  invoiceNumber: string
  dueDate: string
  total: string
  balanceDue: number
  effectiveStatus: string
  client?: { name: string }
}

const { data: invoicesData, pending: isLoading, refresh } = useFetch<Invoice[]>('/api/invoices')
const invoices = computed(() => invoicesData.value ?? [])

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

defineExpose({ refresh })
</script>

<template>
  <div class="rounded-md border">
    <Table>
      <TableHeader>
        <TableRow>
          <TableHead>Invoice #</TableHead>
          <TableHead>Client</TableHead>
          <TableHead>Due Date</TableHead>
          <TableHead>Total</TableHead>
          <TableHead>Balance Due</TableHead>
          <TableHead>Status</TableHead>
          <TableHead class="text-right">
            Actions
          </TableHead>
        </TableRow>
      </TableHeader>
      <TableBody v-if="isLoading">
        <TableRow v-for="i in 5" :key="`skeleton-${i}`">
          <TableCell colspan="7">
            <Skeleton class="h-6 w-full" />
          </TableCell>
        </TableRow>
      </TableBody>
      <TableBody v-else>
        <TableRow v-for="invoice in invoices" :key="invoice.id">
          <TableCell class="font-medium">
            {{ invoice.invoiceNumber }}
          </TableCell>
          <TableCell>{{ invoice.client?.name || 'N/A' }}</TableCell>
          <TableCell>{{ invoice.dueDate }}</TableCell>
          <TableCell>{{ Number(invoice.total).toFixed(2) }}</TableCell>
          <TableCell>{{ invoice.balanceDue.toFixed(2) }}</TableCell>
          <TableCell>
            <Badge :variant="getBadgeVariant(invoice.effectiveStatus)">
              {{ invoice.effectiveStatus }}
            </Badge>
          </TableCell>
          <TableCell class="text-right">
            <NuxtLink :to="`/finance/invoices/${invoice.id}`">
              <Button size="sm" variant="outline">
                View
              </Button>
            </NuxtLink>
          </TableCell>
        </TableRow>
        <TableRow v-if="invoices.length === 0">
          <TableCell colspan="7" class="text-center text-muted-foreground">
            No invoices yet.
          </TableCell>
        </TableRow>
      </TableBody>
    </Table>
  </div>
</template>
