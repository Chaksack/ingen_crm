import { eq } from 'drizzle-orm'
import { db } from '../../db/client'
import { vendors } from '../../db/schema'

export default defineEventHandler(async (event) => {
  await requireUser(event)
  const id = getRouterParam(event, 'id')!
  await db.delete(vendors).where(eq(vendors.id, id))
  return { ok: true }
})
