import { eq } from 'drizzle-orm'
import { z } from 'zod'
import { db } from '../../../db/client'
import { taxRates } from '../../../db/schema'

const bodySchema = z.object({
  name: z.string().min(1).optional(),
  rate: z.number().min(0).optional(),
  compound: z.boolean().optional(),
  order: z.number().optional(),
  active: z.boolean().optional(),
})

export default defineEventHandler(async (event) => {
  await requireUser(event)
  const id = getRouterParam(event, 'id')!
  const body = await readValidatedBody(event, bodySchema.parse)

  const [rate] = await db.update(taxRates).set({
    ...body,
    rate: body.rate !== undefined ? String(body.rate) : undefined,
  }).where(eq(taxRates.id, id)).returning()

  if (!rate) {
    throw createError({ statusCode: 404, statusMessage: 'Tax rate not found' })
  }
  return rate
})
