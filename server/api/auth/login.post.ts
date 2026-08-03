import { eq } from 'drizzle-orm'
import { z } from 'zod'
import { db } from '../../db/client'
import { authCodes, users } from '../../db/schema'

const bodySchema = z.object({
  email: z.string().email(),
  password: z.string().min(1),
})

export default defineEventHandler(async (event) => {
  const body = await readValidatedBody(event, bodySchema.parse)

  const user = await db.query.users.findFirst({ where: eq(users.email, body.email) })
  if (!user || user.status !== 'active' || !user.passwordHash) {
    throw createError({ statusCode: 401, statusMessage: 'Invalid email or password' })
  }

  const valid = await verifyPassword(body.password, user.passwordHash)
  if (!valid) {
    throw createError({ statusCode: 401, statusMessage: 'Invalid email or password' })
  }

  const code = generateCode()
  await db.insert(authCodes).values({
    userId: user.id,
    code,
    purpose: 'login_otp',
    expiresAt: new Date(Date.now() + 10 * 60 * 1000),
  })

  await setPendingAuthCookie(event, user.id, 'login')
  await sendLoginOtpEmail(user.email, code)

  return { ok: true }
})
