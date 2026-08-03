import { eq } from 'drizzle-orm'
import { z } from 'zod'
import { db } from '../../db/client'
import { clients } from '../../db/schema'

const bodySchema = z.object({
  name: z.string().min(1).optional(),
  contactPerson: z.string().optional(),
  email: z.string().email().optional(),
  phone: z.string().optional(),
  address: z.string().optional(),
  notes: z.string().optional(),
  status: z.enum(['active', 'inactive']).optional(),
  creditScore: z.number().optional(),
  kyc: z.boolean().optional(),
  monthlyIncome: z.union([z.number(), z.string()]).optional(),
})

export default defineEventHandler(async (event) => {
  await requireUser(event)
  const id = getRouterParam(event, 'id')!
  const body = await readValidatedBody(event, bodySchema.parse)

  const [client] = await db.update(clients).set({
    ...body,
    monthlyIncome: body.monthlyIncome !== undefined ? String(body.monthlyIncome) : undefined,
  }).where(eq(clients.id, id)).returning()

  if (!client) {
    throw createError({ statusCode: 404, statusMessage: 'Client not found' })
  }

  return client
})
