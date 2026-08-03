import { and, eq, gt, isNull } from 'drizzle-orm'
import { z } from 'zod'
import { db } from '../../db/client'
import { authCodes, users } from '../../db/schema'

const bodySchema = z.object({
  code: z.string().min(1),
})

export default defineEventHandler(async (event) => {
  const body = await readValidatedBody(event, bodySchema.parse)

  const pending = await getPendingAuthCookie(event)
  if (!pending) {
    throw createError({ statusCode: 400, statusMessage: 'No pending verification. Please sign in again.' })
  }

  const purpose = pending.purpose === 'login' ? 'login_otp' : 'password_reset'

  const match = await db.query.authCodes.findFirst({
    where: and(
      eq(authCodes.userId, pending.sub),
      eq(authCodes.purpose, purpose),
      eq(authCodes.code, body.code),
      isNull(authCodes.consumedAt),
      gt(authCodes.expiresAt, new Date()),
    ),
    orderBy: (fields, { desc }) => desc(fields.createdAt),
  })

  if (!match) {
    throw createError({ statusCode: 400, statusMessage: 'Invalid or expired code' })
  }

  await db.update(authCodes).set({ consumedAt: new Date() }).where(eq(authCodes.id, match.id))

  const user = await db.query.users.findFirst({ where: eq(users.id, pending.sub) })
  if (!user) {
    throw createError({ statusCode: 404, statusMessage: 'User not found' })
  }

  clearPendingAuthCookie(event)

  if (pending.purpose === 'login') {
    await setSessionCookie(event, user)
    return { purpose: 'login' as const }
  }

  await setResetVerifiedCookie(event, user.id)
  return { purpose: 'password_reset' as const }
})
