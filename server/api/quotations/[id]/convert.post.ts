import { asc, eq } from 'drizzle-orm'
import { z } from 'zod'
import { db } from '../../../db/client'
import { invoiceLineItems, invoices, invoiceTaxes, quotationLineItems, quotations, quotationTaxes } from '../../../db/schema'

const bodySchema = z.object({
  issueDate: z.string().min(1),
  dueDate: z.string().min(1),
})

export default defineEventHandler(async (event) => {
  const user = await requireUser(event)
  const id = getRouterParam(event, 'id')!
  const body = await readValidatedBody(event, bodySchema.parse)

  const quotation = await db.query.quotations.findFirst({ where: eq(quotations.id, id) })
  if (!quotation) {
    throw createError({ statusCode: 404, statusMessage: 'Quotation not found' })
  }
  if (quotation.convertedInvoiceId) {
    throw createError({ statusCode: 409, statusMessage: 'Quotation already converted to an invoice' })
  }

  const [lineItems, taxes] = await Promise.all([
    db.select().from(quotationLineItems).where(eq(quotationLineItems.quotationId, id)).orderBy(asc(quotationLineItems.order)),
    db.select().from(quotationTaxes).where(eq(quotationTaxes.quotationId, id)).orderBy(asc(quotationTaxes.order)),
  ])

  const invoiceNumber = await nextDocumentNumber('INV')

  return db.transaction(async (tx) => {
    const [invoice] = await tx.insert(invoices).values({
      invoiceNumber,
      clientId: quotation.clientId,
      issueDate: body.issueDate,
      dueDate: body.dueDate,
      currency: quotation.currency,
      subtotal: quotation.subtotal,
      taxExempt: quotation.taxExempt,
      taxAmount: quotation.taxAmount,
      discount: quotation.discount,
      total: quotation.total,
      status: 'draft',
      notes: quotation.notes,
      terms: quotation.terms,
      createdBy: user.sub,
    }).returning()

    if (lineItems.length) {
      await tx.insert(invoiceLineItems).values(lineItems.map(item => ({
        description: item.description,
        quantity: item.quantity,
        unitPrice: item.unitPrice,
        amount: item.amount,
        order: item.order,
        invoiceId: invoice.id,
      })))
    }

    if (taxes.length) {
      await tx.insert(invoiceTaxes).values(taxes.map(tax => ({
        name: tax.name,
        rate: tax.rate,
        amount: tax.amount,
        compound: tax.compound,
        order: tax.order,
        invoiceId: invoice.id,
      })))
    }

    await tx.update(quotations).set({ status: 'accepted', convertedInvoiceId: invoice.id }).where(eq(quotations.id, id))

    return invoice
  })
})
