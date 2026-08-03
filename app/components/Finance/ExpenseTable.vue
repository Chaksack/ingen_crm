<script setup lang="ts">
interface Expense {
  id: string
  category: string
  amount: string
  currency: string
  expenseDate: string
  paymentMethod: string
  status: string
  receiptUrl?: string | null
}

const { data: expensesData, pending: isLoading, refresh } = useFetch<Expense[]>('/api/expenses')
const expenses = computed(() => expensesData.value ?? [])

defineExpose({ refresh })
</script>

<template>
  <div class="rounded-md border">
    <Table>
      <TableHeader>
        <TableRow>
          <TableHead>Category</TableHead>
          <TableHead>Amount</TableHead>
          <TableHead>Date</TableHead>
          <TableHead>Method</TableHead>
          <TableHead>Receipt</TableHead>
        </TableRow>
      </TableHeader>
      <TableBody v-if="isLoading">
        <TableRow v-for="i in 5" :key="`skeleton-${i}`">
          <TableCell colspan="5">
            <Skeleton class="h-6 w-full" />
          </TableCell>
        </TableRow>
      </TableBody>
      <TableBody v-else>
        <TableRow v-for="expense in expenses" :key="expense.id">
          <TableCell>{{ expense.category }}</TableCell>
          <TableCell>{{ Number(expense.amount).toFixed(2) }} {{ expense.currency }}</TableCell>
          <TableCell>{{ expense.expenseDate }}</TableCell>
          <TableCell>{{ expense.paymentMethod }}</TableCell>
          <TableCell>
            <a v-if="expense.receiptUrl" :href="expense.receiptUrl" target="_blank" class="text-primary hover:underline">
              View
            </a>
            <span v-else>-</span>
          </TableCell>
        </TableRow>
        <TableRow v-if="expenses.length === 0">
          <TableCell colspan="5" class="text-center text-muted-foreground">
            No expenses recorded yet.
          </TableCell>
        </TableRow>
      </TableBody>
    </Table>
  </div>
</template>
