import { eq } from 'drizzle-orm'
import { z } from 'zod'
import { db } from '../../db/client'
import { businesses } from '../../db/schema'

const bodySchema = z.object({
  companyName: z.string().min(1).optional(),
  companyNumber: z.string().optional(),
  email: z.string().email().optional(),
  phoneNumber: z.string().optional(),
  address: z.string().optional(),
  avatar: z.string().optional(),
  capital: z.union([z.number(), z.string()]).optional(),
  foundedYear: z.number().optional(),
  country: z.string().optional(),
  currency: z.string().optional(),
  creditScore: z.number().optional(),
  kyc: z.boolean().optional(),
  status: z.enum(['active', 'inactive', 'pending']).optional(),
})

export default defineEventHandler(async (event) => {
  await requireUser(event)
  const id = getRouterParam(event, 'id')!
  const body = await readValidatedBody(event, bodySchema.parse)

  const [row] = await db.update(businesses).set({
    ...body,
    capital: body.capital !== undefined ? String(body.capital) : undefined,
  }).where(eq(businesses.id, id)).returning()

  if (!row) {
    throw createError({ statusCode: 404, statusMessage: 'Business not found' })
  }

  return row
})
