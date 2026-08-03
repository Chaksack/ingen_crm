<script setup lang="ts">
interface FinanceDocument {
  id: string
  documentType: 'invoice' | 'quotation'
  number: string
  date: string
  total: string
  status: string
  client?: { name: string }
}

const { data: documentsData, pending: isLoading, refresh } = useFetch<FinanceDocument[]>('/api/finance/documents')
const documents = computed(() => documentsData.value ?? [])

const typeFilter = ref<'all' | 'invoice' | 'quotation'>('all')

const filtered = computed(() => {
  if (typeFilter.value === 'all')
    return documents.value
  return documents.value.filter(d => d.documentType === typeFilter.value)
})

function getBadgeVariant(doc: FinanceDocument) {
  if (doc.documentType === 'invoice') {
    switch (doc.status) {
      case 'paid': return 'default'
      case 'overdue': return 'destructive'
      default: return 'secondary'
    }
  }
  switch (doc.status) {
    case 'accepted': return 'default'
    case 'rejected':
    case 'expired': return 'destructive'
    default: return 'secondary'
  }
}

function detailLink(doc: FinanceDocument) {
  return doc.documentType === 'invoice' ? `/finance/invoices/${doc.id}` : `/finance/quotations/${doc.id}`
}

defineExpose({ refresh })
</script>

<template>
  <div class="space-y-4">
    <Tabs v-model="typeFilter">
      <TabsList>
        <TabsTrigger value="all">
          All
        </TabsTrigger>
        <TabsTrigger value="invoice">
          Invoices
        </TabsTrigger>
        <TabsTrigger value="quotation">
          Quotations
        </TabsTrigger>
      </TabsList>
    </Tabs>

    <div class="rounded-md border">
      <Table>
        <TableHeader>
          <TableRow>
            <TableHead>Type</TableHead>
            <TableHead>Number</TableHead>
            <TableHead>Client</TableHead>
            <TableHead>Date</TableHead>
            <TableHead>Total</TableHead>
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
          <TableRow v-for="doc in filtered" :key="`${doc.documentType}-${doc.id}`">
            <TableCell>
              <Badge variant="outline">
                {{ doc.documentType === 'invoice' ? 'Invoice' : 'Quotation' }}
              </Badge>
            </TableCell>
            <TableCell class="font-medium">
              {{ doc.number }}
            </TableCell>
            <TableCell>{{ doc.client?.name || 'N/A' }}</TableCell>
            <TableCell>{{ doc.date }}</TableCell>
            <TableCell>{{ Number(doc.total).toFixed(2) }}</TableCell>
            <TableCell>
              <Badge :variant="getBadgeVariant(doc)">
                {{ doc.status }}
              </Badge>
            </TableCell>
            <TableCell class="text-right">
              <NuxtLink :to="detailLink(doc)">
                <Button size="sm" variant="outline">
                  View
                </Button>
              </NuxtLink>
            </TableCell>
          </TableRow>
          <TableRow v-if="filtered.length === 0">
            <TableCell colspan="7" class="text-center text-muted-foreground">
              No documents yet.
            </TableCell>
          </TableRow>
        </TableBody>
      </Table>
    </div>
  </div>
</template>
