import { desc, isNotNull } from 'drizzle-orm'
import { db } from '../../db/client'
import { staff, users } from '../../db/schema'

export default defineEventHandler(async (event) => {
  await requireUser(event)

  const [staffRows, linkedUsers] = await Promise.all([
    db.select().from(staff).orderBy(desc(staff.createdAt)),
    db.select({ staffId: users.staffId, status: users.status }).from(users).where(isNotNull(users.staffId)),
  ])

  const loginStatusByStaffId = new Map(linkedUsers.map(u => [u.staffId, u.status]))

  return staffRows.map(row => ({
    ...row,
    loginStatus: loginStatusByStaffId.get(row.id) ?? null,
  }))
})
