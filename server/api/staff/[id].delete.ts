import { eq } from 'drizzle-orm'
import { db } from '../../db/client'
import { staff } from '../../db/schema'

export default defineEventHandler(async (event) => {
  await requireUser(event)
  const id = getRouterParam(event, 'id')!
  await db.delete(staff).where(eq(staff.id, id))
  return { ok: true }
})
