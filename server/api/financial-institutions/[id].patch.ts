import { eq } from 'drizzle-orm'
import { z } from 'zod'
import { db } from '../../db/client'
import { financialInstitutions } from '../../db/schema'

const bodySchema = z.object({
  institutionName: z.string().min(1).optional(),
  email: z.string().email().optional(),
  phone: z.string().optional(),
  address: z.string().optional(),
  website: z.string().optional(),
  avatar: z.string().optional(),
  institutionType: z.string().optional(),
  registrationNumber: z.string().optional(),
  foundedDate: z.string().optional(),
  country: z.string().optional(),
  services: z.string().optional(),
  description: z.string().optional(),
  status: z.enum(['active', 'inactive', 'pending']).optional(),
})

export default defineEventHandler(async (event) => {
  await requireUser(event)
  const id = getRouterParam(event, 'id')!
  const body = await readValidatedBody(event, bodySchema.parse)
  const [row] = await db.update(financialInstitutions).set(body).where(eq(financialInstitutions.id, id)).returning()
  if (!row) {
    throw createError({ statusCode: 404, statusMessage: 'Financial institution not found' })
  }
  return row
})
