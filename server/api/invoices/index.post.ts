import { z } from 'zod'
import { db } from '../../db/client'
import { invoiceLineItems, invoices, invoiceTaxes } from '../../db/schema'

const bodySchema = z.object({
  clientId: z.string().min(1),
  projectId: z.string().optional(),
  issueDate: z.string().min(1),
  dueDate: z.string().min(1),
  currency: z.string().optional(),
  taxExempt: z.boolean().optional(),
  discount: z.number().optional(),
  notes: z.string().optional(),
  terms: z.string().optional(),
  status: z.enum(['draft', 'sent']).optional(),
  lineItems: z.array(z.object({
    description: z.string().min(1),
    quantity: z.number().positive(),
    unitPrice: z.number().nonnegative(),
  })).min(1),
})

export default defineEventHandler(async (event) => {
  const user = await requireUser(event)
  const body = await readValidatedBody(event, bodySchema.parse)

  const items = computeLineItems(body.lineItems)
  const { subtotal, taxableBase } = computeTotals(items, body.discount ?? 0)
  const rates = body.taxExempt ? [] : await getActiveTaxRates()
  const { lines: taxLines, taxAmount } = computeTaxLines(taxableBase, rates)
  const total = Math.round((taxableBase + taxAmount) * 100) / 100
  const invoiceNumber = await nextDocumentNumber('INV')

  return db.transaction(async (tx) => {
    const [invoice] = await tx.insert(invoices).values({
      invoiceNumber,
      clientId: body.clientId,
      projectId: body.projectId,
      issueDate: body.issueDate,
      dueDate: body.dueDate,
      currency: body.currency ?? 'GHS',
      subtotal: String(subtotal),
      taxExempt: body.taxExempt ?? false,
      taxAmount: String(taxAmount),
      discount: String(body.discount ?? 0),
      total: String(total),
      status: body.status ?? 'draft',
      notes: body.notes,
      terms: body.terms,
      createdBy: user.sub,
    }).returning()

    await tx.insert(invoiceLineItems).values(items.map(item => ({ ...item, invoiceId: invoice.id })))
    if (taxLines.length) {
      await tx.insert(invoiceTaxes).values(taxLines.map(line => ({
        name: line.name,
        rate: String(line.rate),
        amount: String(line.amount),
        compound: line.compound,
        order: line.order,
        invoiceId: invoice.id,
      })))
    }

    return invoice
  })
})
