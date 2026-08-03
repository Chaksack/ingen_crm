import { eq } from 'drizzle-orm'
import { z } from 'zod'
import { db } from '../../db/client'
import { staff } from '../../db/schema'

const bodySchema = z.object({
  firstName: z.string().min(1).optional(),
  lastName: z.string().min(1).optional(),
  email: z.string().email().optional(),
  phoneNumber: z.string().optional(),
  role: z.enum(['Staff', 'Manager', 'Admin']).optional(),
  department: z.string().optional(),
  status: z.enum(['active', 'inactive']).optional(),
  avatar: z.string().optional(),
})

export default defineEventHandler(async (event) => {
  await requireUser(event)
  const id = getRouterParam(event, 'id')!
  const body = await readValidatedBody(event, bodySchema.parse)
  const [row] = await db.update(staff).set(body).where(eq(staff.id, id)).returning()
  if (!row) {
    throw createError({ statusCode: 404, statusMessage: 'Staff not found' })
  }
  return row
})
