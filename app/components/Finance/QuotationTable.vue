<script setup lang="ts">
interface Quotation {
  id: string
  quoteNumber: string
  expiryDate: string
  total: string
  status: string
  convertedInvoiceId?: string | null
  client?: { name: string }
}

const { data: quotationsData, pending: isLoading, refresh } = useFetch<Quotation[]>('/api/quotations')
const quotations = computed(() => quotationsData.value ?? [])

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

defineExpose({ refresh })
</script>

<template>
  <div class="rounded-md border">
    <Table>
      <TableHeader>
        <TableRow>
          <TableHead>Quote #</TableHead>
          <TableHead>Client</TableHead>
          <TableHead>Expiry</TableHead>
          <TableHead>Total</TableHead>
          <TableHead>Status</TableHead>
          <TableHead class="text-right">
            Actions
          </TableHead>
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
        <TableRow v-for="quotation in quotations" :key="quotation.id">
          <TableCell class="font-medium">
            {{ quotation.quoteNumber }}
          </TableCell>
          <TableCell>{{ quotation.client?.name || 'N/A' }}</TableCell>
          <TableCell>{{ quotation.expiryDate }}</TableCell>
          <TableCell>{{ Number(quotation.total).toFixed(2) }}</TableCell>
          <TableCell>
            <Badge :variant="getBadgeVariant(quotation.status)">
              {{ quotation.status }}
            </Badge>
          </TableCell>
          <TableCell class="text-right">
            <NuxtLink :to="`/finance/quotations/${quotation.id}`">
              <Button size="sm" variant="outline">
                View
              </Button>
            </NuxtLink>
          </TableCell>
        </TableRow>
        <TableRow v-if="quotations.length === 0">
          <TableCell colspan="6" class="text-center text-muted-foreground">
            No quotations yet.
          </TableCell>
        </TableRow>
      </TableBody>
    </Table>
  </div>
</template>
