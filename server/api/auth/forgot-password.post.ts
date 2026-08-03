import { eq } from 'drizzle-orm'
import { z } from 'zod'
import { db } from '../../db/client'
import { authCodes, users } from '../../db/schema'

const bodySchema = z.object({
  email: z.string().email(),
})

export default defineEventHandler(async (event) => {
  const body = await readValidatedBody(event, bodySchema.parse)

  const user = await db.query.users.findFirst({ where: eq(users.email, body.email) })

  // Always respond ok to avoid leaking whether an email is registered.
  if (!user) {
    return { ok: true }
  }

  const code = generateCode()
  await db.insert(authCodes).values({
    userId: user.id,
    code,
    purpose: 'password_reset',
    expiresAt: new Date(Date.now() + 10 * 60 * 1000),
  })

  await setPendingAuthCookie(event, user.id, 'reset')
  await sendPasswordResetEmail(user.email, code)

  return { ok: true }
})
