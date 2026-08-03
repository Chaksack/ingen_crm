import { eq } from 'drizzle-orm'
import { db } from '../../db/client'
import { staff, users } from '../../db/schema'

export default defineEventHandler(async (event) => {
  const session = await requireUser(event)
  if (session.role !== 'admin') {
    throw createError({ statusCode: 403, statusMessage: 'Only an admin can delete staff members' })
  }

  const id = getRouterParam(event, 'id')!

  await db.transaction(async (tx) => {
    // Revoke any linked login rather than leaving it active and orphaned once the
    // staff row (and its FK, via ON DELETE SET NULL) is gone.
    await tx.update(users).set({ status: 'inactive' }).where(eq(users.staffId, id))
    await tx.delete(staff).where(eq(staff.id, id))
  })

  return { ok: true }
})
