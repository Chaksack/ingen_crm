import { eq } from 'drizzle-orm'
import { z } from 'zod'
import { db } from '../../db/client'
import { vendors } from '../../db/schema'

const bodySchema = z.object({
  vendorName: z.string().min(1).optional(),
  email: z.string().email().optional(),
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
  const id = getRouterParam(event, 'id')!
  const body = await readValidatedBody(event, bodySchema.parse)
  const [row] = await db.update(vendors).set(body).where(eq(vendors.id, id)).returning()
  if (!row) {
    throw createError({ statusCode: 404, statusMessage: 'Vendor not found' })
  }
  return row
})
