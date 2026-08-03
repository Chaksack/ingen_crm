import { eq } from 'drizzle-orm'
import { db } from '../../db/client'
import { quotations } from '../../db/schema'

export default defineEventHandler(async (event) => {
  await requireUser(event)
  const id = getRouterParam(event, 'id')!
  await db.delete(quotations).where(eq(quotations.id, id))
  return { ok: true }
})
