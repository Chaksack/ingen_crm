import { asc, eq } from 'drizzle-orm'
import { db } from '../../db/client'
import { invoiceLineItems, invoices, invoiceTaxes, payments } from '../../db/schema'

export default defineEventHandler(async (event) => {
  await requireUser(event)
  const id = getRouterParam(event, 'id')!

  const invoice = await db.query.invoices.findFirst({
    where: eq(invoices.id, id),
    with: { client: true },
  })
  if (!invoice) {
    throw createError({ statusCode: 404, statusMessage: 'Invoice not found' })
  }

  const [lineItems, taxes, invoicePayments] = await Promise.all([
    db.select().from(invoiceLineItems).where(eq(invoiceLineItems.invoiceId, id)).orderBy(asc(invoiceLineItems.order)),
    db.select().from(invoiceTaxes).where(eq(invoiceTaxes.invoiceId, id)).orderBy(asc(invoiceTaxes.order)),
    db.select().from(payments).where(eq(payments.invoiceId, id)).orderBy(asc(payments.paidAt)),
  ])

  return { ...computeInvoiceView(invoice, invoicePayments), lineItems, taxes, payments: invoicePayments }
})
