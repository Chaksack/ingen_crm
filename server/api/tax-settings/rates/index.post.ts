import { z } from 'zod'
import { db } from '../../../db/client'
import { taxRates } from '../../../db/schema'

const bodySchema = z.object({
  name: z.string().min(1),
  rate: z.number().min(0),
  compound: z.boolean().optional(),
  order: z.number().optional(),
  active: z.boolean().optional(),
})

export default defineEventHandler(async (event) => {
  await requireUser(event)
  const body = await readValidatedBody(event, bodySchema.parse)

  const [rate] = await db.insert(taxRates).values({
    name: body.name,
    rate: String(body.rate),
    compound: body.compound ?? false,
    order: body.order ?? 0,
    active: body.active ?? true,
  }).returning()

  return rate
})
