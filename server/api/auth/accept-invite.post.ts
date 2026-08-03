import { and, eq, gt, isNull } from 'drizzle-orm'
import { z } from 'zod'
import { db } from '../../db/client'
import { authCodes, users } from '../../db/schema'

const bodySchema = z.object({
  token: z.string().min(1),
  password: z.string().min(6),
})

export default defineEventHandler(async (event) => {
  const body = await readValidatedBody(event, bodySchema.parse)

  const invite = await db.query.authCodes.findFirst({
    where: and(
      eq(authCodes.code, body.token),
      eq(authCodes.purpose, 'invite'),
      isNull(authCodes.consumedAt),
      gt(authCodes.expiresAt, new Date()),
    ),
  })

  if (!invite) {
    throw createError({ statusCode: 400, statusMessage: 'This invite link is invalid or has expired.' })
  }

  const passwordHash = await hashPassword(body.password)

  const [user] = await db.update(users)
    .set({ passwordHash, status: 'active' })
    .where(eq(users.id, invite.userId))
    .returning()

  await db.update(authCodes).set({ consumedAt: new Date() }).where(eq(authCodes.id, invite.id))

  return { ok: true, email: user.email }
})
