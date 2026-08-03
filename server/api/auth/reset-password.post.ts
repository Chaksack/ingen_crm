import { eq } from 'drizzle-orm'
import { z } from 'zod'
import { db } from '../../db/client'
import { users } from '../../db/schema'

const bodySchema = z.object({
  password: z.string().min(6),
})

export default defineEventHandler(async (event) => {
  const body = await readValidatedBody(event, bodySchema.parse)

  const verified = await getResetVerifiedCookie(event)
  if (!verified) {
    throw createError({ statusCode: 400, statusMessage: 'Password reset not verified. Please start over.' })
  }

  const passwordHash = await hashPassword(body.password)
  await db.update(users).set({ passwordHash }).where(eq(users.id, verified.sub))

  clearResetVerifiedCookie(event)

  return { ok: true }
})
