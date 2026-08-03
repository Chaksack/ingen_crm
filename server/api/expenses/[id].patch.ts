import { eq } from 'drizzle-orm'
import { z } from 'zod'
import { db } from '../../db/client'
import { expenses } from '../../db/schema'

const bodySchema = z.object({
  category: z.string().min(1).optional(),
  vendorId: z.string().optional(),
  amount: z.number().positive().optional(),
  currency: z.string().optional(),
  expenseDate: z.string().optional(),
  description: z.string().optional(),
  paymentMethod: z.enum(['cash', 'bank_transfer', 'mobile_money', 'card', 'cheque']).optional(),
  receiptUrl: z.string().optional(),
  status: z.enum(['recorded', 'approved', 'rejected']).optional(),
})

export default defineEventHandler(async (event) => {
  await requireUser(event)
  const id = getRouterParam(event, 'id')!
  const body = await readValidatedBody(event, bodySchema.parse)

  const [row] = await db.update(expenses).set({
    ...body,
    amount: body.amount !== undefined ? String(body.amount) : undefined,
  }).where(eq(expenses.id, id)).returning()

  if (!row) {
    throw createError({ statusCode: 404, statusMessage: 'Expense not found' })
  }
  return row
})
