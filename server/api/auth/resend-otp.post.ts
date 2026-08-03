import { eq } from 'drizzle-orm'
import { db } from '../../db/client'
import { authCodes, users } from '../../db/schema'

export default defineEventHandler(async (event) => {
  const pending = await getPendingAuthCookie(event)
  if (!pending) {
    throw createError({ statusCode: 400, statusMessage: 'No pending verification. Please start over.' })
  }

  const user = await db.query.users.findFirst({ where: eq(users.id, pending.sub) })
  if (!user) {
    throw createError({ statusCode: 404, statusMessage: 'User not found' })
  }

  const code = generateCode()
  const purpose = pending.purpose === 'login' ? 'login_otp' : 'password_reset'
  await db.insert(authCodes).values({
    userId: user.id,
    code,
    purpose,
    expiresAt: new Date(Date.now() + 10 * 60 * 1000),
  })

  if (pending.purpose === 'login')
    await sendLoginOtpEmail(user.email, code)
  else
    await sendPasswordResetEmail(user.email, code)

  return { ok: true }
})
