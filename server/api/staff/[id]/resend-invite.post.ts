import { and, eq, isNull } from 'drizzle-orm'
import { db } from '../../../db/client'
import { authCodes, staff, users } from '../../../db/schema'

export default defineEventHandler(async (event) => {
  const session = await requireUser(event)
  if (session.role !== 'admin') {
    throw createError({ statusCode: 403, statusMessage: 'Only an admin can manage staff logins' })
  }

  const staffId = getRouterParam(event, 'id')!
  const staffRow = await db.query.staff.findFirst({ where: eq(staff.id, staffId) })
  if (!staffRow) {
    throw createError({ statusCode: 404, statusMessage: 'Staff member not found' })
  }

  const user = await db.query.users.findFirst({ where: eq(users.staffId, staffId) })
  if (!user) {
    throw createError({ statusCode: 404, statusMessage: 'This staff member has no login to invite' })
  }
  if (user.status !== 'pending') {
    throw createError({ statusCode: 409, statusMessage: 'This staff member has already activated their account' })
  }

  // Invalidate any earlier unconsumed invite codes for this user before issuing a new one.
  await db.update(authCodes)
    .set({ consumedAt: new Date() })
    .where(and(eq(authCodes.userId, user.id), eq(authCodes.purpose, 'invite'), isNull(authCodes.consumedAt)))

  const token = generateToken()
  await db.insert(authCodes).values({
    userId: user.id,
    code: token,
    purpose: 'invite',
    expiresAt: new Date(Date.now() + 7 * 24 * 60 * 60 * 1000),
  })

  const inviteLink = `${getRequestURL(event).origin}/accept-invite?token=${token}`
  const inviteEmailSent = await sendStaffInviteEmail(user.email, user.name, inviteLink)

  return {
    inviteEmailSent,
    inviteLink: inviteEmailSent ? null : inviteLink,
  }
})
