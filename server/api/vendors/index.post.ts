import { z } from 'zod'
import { db } from '../../db/client'
import { vendors } from '../../db/schema'

const bodySchema = z.object({
  vendorName: z.string().min(1),
  email: z.string().email(),
  phone: z.string().optional(),
  address: z.string().optional(),
  website: z.string().optional(),
  avatar: z.string().optional(),
  vendorType: z.string().optional(),
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
  const [row] = await db.insert(vendors).values(body).returning()
  return row
})
