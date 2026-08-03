import { and, eq, gt, isNull } from 'drizzle-orm'
import { db } from '../../db/client'
import { authCodes, users } from '../../db/schema'

export default defineEventHandler(async (event) => {
  const query = getQuery(event)
  const token = typeof query.token === 'string' ? query.token : ''

  const invite = await db.query.authCodes.findFirst({
    where: and(
      eq(authCodes.code, token),
      eq(authCodes.purpose, 'invite'),
      isNull(authCodes.consumedAt),
      gt(authCodes.expiresAt, new Date()),
    ),
  })

  if (!invite) {
    throw createError({ statusCode: 400, statusMessage: 'This invite link is invalid or has expired.' })
  }

  const user = await db.query.users.findFirst({ where: eq(users.id, invite.userId) })

  return { email: user?.email, name: user?.name }
})
