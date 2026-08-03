import { eq } from 'drizzle-orm'
import { z } from 'zod'
import { db } from '../../db/client'
import { invoices } from '../../db/schema'

const bodySchema = z.object({
  status: z.enum(['draft', 'sent', 'void']).optional(),
  dueDate: z.string().optional(),
  notes: z.string().optional(),
  terms: z.string().optional(),
})

export default defineEventHandler(async (event) => {
  await requireUser(event)
  const id = getRouterParam(event, 'id')!
  const body = await readValidatedBody(event, bodySchema.parse)

  const [row] = await db.update(invoices).set(body).where(eq(invoices.id, id)).returning()
  if (!row) {
    throw createError({ statusCode: 404, statusMessage: 'Invoice not found' })
  }
  return row
})
