import { eq } from 'drizzle-orm'
import { db } from '../../db/client'
import { users } from '../../db/schema'

export default defineEventHandler(async (event) => {
  const session = await getSessionUser(event)
  if (!session) {
    throw createError({ statusCode: 401, statusMessage: 'Unauthorized' })
  }

  const user = await db.query.users.findFirst({ where: eq(users.id, session.sub) })
  if (!user) {
    throw createError({ statusCode: 401, statusMessage: 'Unauthorized' })
  }

  return {
    id: user.id,
    email: user.email,
    name: user.name,
    avatar: user.avatar,
    role: user.role,
  }
})
