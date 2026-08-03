import { z } from 'zod'
import { db } from '../../db/client'
import { quotationLineItems, quotations, quotationTaxes } from '../../db/schema'

const bodySchema = z.object({
  clientId: z.string().min(1),
  issueDate: z.string().min(1),
  expiryDate: z.string().min(1),
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
  const quoteNumber = await nextDocumentNumber('QUO')

  return db.transaction(async (tx) => {
    const [quotation] = await tx.insert(quotations).values({
      quoteNumber,
      clientId: body.clientId,
      issueDate: body.issueDate,
      expiryDate: body.expiryDate,
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

    await tx.insert(quotationLineItems).values(items.map(item => ({ ...item, quotationId: quotation.id })))
    if (taxLines.length) {
      await tx.insert(quotationTaxes).values(taxLines.map(line => ({
        name: line.name,
        rate: String(line.rate),
        amount: String(line.amount),
        compound: line.compound,
        order: line.order,
        quotationId: quotation.id,
      })))
    }

    return quotation
  })
})
