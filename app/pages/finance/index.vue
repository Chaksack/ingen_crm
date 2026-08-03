<script setup lang="ts">
interface FinanceSummary {
  outstandingBalance: number
  invoicedThisMonth: number
  collectedThisMonth: number
  expensesThisMonth: number
  openQuotations: number
}

const { data: summary, pending: isLoading } = await useFetch<FinanceSummary>('/api/finance/summary')

function formatMoney(value: number | undefined) {
  return new Intl.NumberFormat('en-GH', { style: 'currency', currency: 'GHS' }).format(value ?? 0)
}
</script>

<template>
  <FinanceLayout>
    <div class="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
      <Card>
        <CardHeader class="pb-2">
          <CardDescription>Outstanding Balance</CardDescription>
          <CardTitle class="text-2xl">
            <Skeleton v-if="isLoading" class="h-8 w-32" />
            <span v-else>{{ formatMoney(summary?.outstandingBalance) }}</span>
          </CardTitle>
        </CardHeader>
      </Card>
      <Card>
        <CardHeader class="pb-2">
          <CardDescription>Invoiced This Month</CardDescription>
          <CardTitle class="text-2xl">
            <Skeleton v-if="isLoading" class="h-8 w-32" />
            <span v-else>{{ formatMoney(summary?.invoicedThisMonth) }}</span>
          </CardTitle>
        </CardHeader>
      </Card>
      <Card>
        <CardHeader class="pb-2">
          <CardDescription>Collected This Month</CardDescription>
          <CardTitle class="text-2xl">
            <Skeleton v-if="isLoading" class="h-8 w-32" />
            <span v-else>{{ formatMoney(summary?.collectedThisMonth) }}</span>
          </CardTitle>
        </CardHeader>
      </Card>
      <Card>
        <CardHeader class="pb-2">
          <CardDescription>Expenses This Month</CardDescription>
          <CardTitle class="text-2xl">
            <Skeleton v-if="isLoading" class="h-8 w-32" />
            <span v-else>{{ formatMoney(summary?.expensesThisMonth) }}</span>
          </CardTitle>
        </CardHeader>
      </Card>
      <Card>
        <CardHeader class="pb-2">
          <CardDescription>Open Quotations</CardDescription>
          <CardTitle class="text-2xl">
            <Skeleton v-if="isLoading" class="h-8 w-16" />
            <span v-else>{{ summary?.openQuotations ?? 0 }}</span>
          </CardTitle>
        </CardHeader>
      </Card>
    </div>
  </FinanceLayout>
</template>
