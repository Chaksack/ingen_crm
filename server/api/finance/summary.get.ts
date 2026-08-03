import { db } from '../../db/client'
import { expenses, invoices, payments, quotations } from '../../db/schema'

function isThisMonth(dateStr: string) {
  const d = new Date(dateStr)
  const now = new Date()
  return d.getFullYear() === now.getFullYear() && d.getMonth() === now.getMonth()
}

export default defineEventHandler(async (event) => {
  await requireUser(event)

  const [allInvoices, allPayments, allExpenses, allQuotations] = await Promise.all([
    db.select().from(invoices),
    db.select().from(payments),
    db.select().from(expenses),
    db.select().from(quotations),
  ])

  const paymentsByInvoice = new Map<string, number>()
  for (const p of allPayments) {
    paymentsByInvoice.set(p.invoiceId, (paymentsByInvoice.get(p.invoiceId) ?? 0) + Number(p.amount))
  }

  const outstandingBalance = allInvoices
    .filter(i => i.status !== 'void' && i.status !== 'draft')
    .reduce((sum, i) => sum + Math.max(Number(i.total) - (paymentsByInvoice.get(i.id) ?? 0), 0), 0)

  const invoicedThisMonth = allInvoices
    .filter(i => i.status !== 'void' && isThisMonth(i.issueDate))
    .reduce((sum, i) => sum + Number(i.total), 0)

  const collectedThisMonth = allPayments
    .filter(p => isThisMonth(p.paidAt))
    .reduce((sum, p) => sum + Number(p.amount), 0)

  const expensesThisMonth = allExpenses
    .filter(e => isThisMonth(e.expenseDate))
    .reduce((sum, e) => sum + Number(e.amount), 0)

  const openQuotations = allQuotations.filter(q => q.status === 'draft' || q.status === 'sent').length

  return {
    outstandingBalance,
    invoicedThisMonth,
    collectedThisMonth,
    expensesThisMonth,
    openQuotations,
  }
})
