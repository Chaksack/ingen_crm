import { eq } from 'drizzle-orm'
import { db } from '../../db/client'
import { clients } from '../../db/schema'

export default defineEventHandler(async (event) => {
  await requireUser(event)
  const id = getRouterParam(event, 'id')!
  await db.delete(clients).where(eq(clients.id, id))
  return { ok: true }
})
