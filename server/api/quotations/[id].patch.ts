import { eq } from 'drizzle-orm'
import { z } from 'zod'
import { db } from '../../db/client'
import { quotations } from '../../db/schema'

const bodySchema = z.object({
  status: z.enum(['draft', 'sent', 'accepted', 'rejected', 'expired']).optional(),
  expiryDate: z.string().optional(),
  notes: z.string().optional(),
  terms: z.string().optional(),
})

export default defineEventHandler(async (event) => {
  await requireUser(event)
  const id = getRouterParam(event, 'id')!
  const body = await readValidatedBody(event, bodySchema.parse)

  const [row] = await db.update(quotations).set(body).where(eq(quotations.id, id)).returning()
  if (!row) {
    throw createError({ statusCode: 404, statusMessage: 'Quotation not found' })
  }
  return row
})
