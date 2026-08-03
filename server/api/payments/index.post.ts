import { eq } from 'drizzle-orm'
import { z } from 'zod'
import { db } from '../../db/client'
import { invoices, payments } from '../../db/schema'

const bodySchema = z.object({
  invoiceId: z.string().min(1),
  amount: z.number().positive(),
  method: z.enum(['cash', 'bank_transfer', 'mobile_money', 'card', 'cheque']),
  reference: z.string().optional(),
  paidAt: z.string().min(1),
  notes: z.string().optional(),
})

export default defineEventHandler(async (event) => {
  const user = await requireUser(event)
  const body = await readValidatedBody(event, bodySchema.parse)

  const invoice = await db.query.invoices.findFirst({ where: eq(invoices.id, body.invoiceId) })
  if (!invoice) {
    throw createError({ statusCode: 404, statusMessage: 'Invoice not found' })
  }

  const [payment] = await db.transaction(async (tx) => {
    const [row] = await tx.insert(payments).values({
      invoiceId: body.invoiceId,
      amount: String(body.amount),
      method: body.method,
      reference: body.reference,
      paidAt: body.paidAt,
      notes: body.notes,
      recordedBy: user.sub,
    }).returning()

    const existingPayments = await tx.select().from(payments).where(eq(payments.invoiceId, body.invoiceId))
    const totalPaid = existingPayments.reduce((sum, p) => sum + Number(p.amount), 0)
    const nextStatus = totalPaid >= Number(invoice.total) ? 'paid' : 'partially_paid'

    if (invoice.status !== 'void') {
      const receiptNumber = invoice.receiptNumber ?? await nextDocumentNumber('RCT')
      await tx.update(invoices).set({ status: nextStatus, receiptNumber }).where(eq(invoices.id, body.invoiceId))
    }

    return [row]
  })

  return payment
})
