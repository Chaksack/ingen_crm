import { z } from 'zod'
import { db } from '../../db/client'
import { financialInstitutions } from '../../db/schema'

const bodySchema = z.object({
  institutionName: z.string().min(1),
  email: z.string().email(),
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
  const body = await readValidatedBody(event, bodySchema.parse)
  const [row] = await db.insert(financialInstitutions).values(body).returning()
  return row
})
