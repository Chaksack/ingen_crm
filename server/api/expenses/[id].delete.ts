import { eq } from 'drizzle-orm'
import { db } from '../../db/client'
import { expenses } from '../../db/schema'

export default defineEventHandler(async (event) => {
  await requireUser(event)
  const id = getRouterParam(event, 'id')!
  await db.delete(expenses).where(eq(expenses.id, id))
  return { ok: true }
})
