import { desc, inArray } from 'drizzle-orm'
import { db } from '../../db/client'
import { invoices, payments, quotations } from '../../db/schema'
import { computeInvoiceView } from '../../utils/invoice'

export default defineEventHandler(async (event) => {
  await requireUser(event)

  const [invoiceRows, quotationRows] = await Promise.all([
    db.query.invoices.findMany({ orderBy: desc(invoices.createdAt), with: { client: true } }),
    db.query.quotations.findMany({ orderBy: desc(quotations.createdAt), with: { client: true } }),
  ])

  const allPayments = invoiceRows.length
    ? await db.select().from(payments).where(inArray(payments.invoiceId, invoiceRows.map(r => r.id)))
    : []
  const paymentsByInvoice = new Map<string, typeof allPayments>()
  for (const p of allPayments) {
    const list = paymentsByInvoice.get(p.invoiceId) ?? []
    list.push(p)
    paymentsByInvoice.set(p.invoiceId, list)
  }

  const documents = [
    ...invoiceRows.map((row) => {
      const view = computeInvoiceView(row, paymentsByInvoice.get(row.id) ?? [])
      return {
        documentType: 'invoice' as const,
        id: row.id,
        number: row.invoiceNumber,
        client: row.client,
        date: row.dueDate,
        total: row.total,
        status: view.effectiveStatus,
        createdAt: row.createdAt,
      }
    }),
    ...quotationRows.map(row => ({
      documentType: 'quotation' as const,
      id: row.id,
      number: row.quoteNumber,
      client: row.client,
      date: row.expiryDate,
      total: row.total,
      status: row.status,
      createdAt: row.createdAt,
    })),
  ]

  documents.sort((a, b) => new Date(b.createdAt).getTime() - new Date(a.createdAt).getTime())

  return documents
})
