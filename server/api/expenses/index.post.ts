import { z } from 'zod'
import { db } from '../../db/client'
import { expenses } from '../../db/schema'

const bodySchema = z.object({
  category: z.string().min(1),
  vendorId: z.string().optional(),
  amount: z.number().positive(),
  currency: z.string().optional(),
  expenseDate: z.string().min(1),
  description: z.string().optional(),
  paymentMethod: z.enum(['cash', 'bank_transfer', 'mobile_money', 'card', 'cheque']),
  receiptUrl: z.string().optional(),
})

export default defineEventHandler(async (event) => {
  const user = await requireUser(event)
  const body = await readValidatedBody(event, bodySchema.parse)

  const [row] = await db.insert(expenses).values({
    ...body,
    amount: String(body.amount),
    recordedBy: user.sub,
  }).returning()

  return row
})
