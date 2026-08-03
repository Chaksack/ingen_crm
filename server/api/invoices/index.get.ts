import { desc, inArray } from 'drizzle-orm'
import { db } from '../../db/client'
import { invoices, payments } from '../../db/schema'

export default defineEventHandler(async (event) => {
  await requireUser(event)

  const rows = await db.query.invoices.findMany({
    orderBy: desc(invoices.createdAt),
    with: { client: true },
  })

  if (!rows.length)
    return []

  const allPayments = await db.select().from(payments).where(inArray(payments.invoiceId, rows.map(r => r.id)))
  const paymentsByInvoice = new Map<string, typeof allPayments>()
  for (const p of allPayments) {
    const list = paymentsByInvoice.get(p.invoiceId) ?? []
    list.push(p)
    paymentsByInvoice.set(p.invoiceId, list)
  }

  return rows.map(row => computeInvoiceView(row, paymentsByInvoice.get(row.id) ?? []))
})
