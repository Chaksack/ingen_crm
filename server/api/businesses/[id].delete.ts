import { eq } from 'drizzle-orm'
import { db } from '../../db/client'
import { businesses } from '../../db/schema'

export default defineEventHandler(async (event) => {
  await requireUser(event)
  const id = getRouterParam(event, 'id')!
  await db.delete(businesses).where(eq(businesses.id, id))
  return { ok: true }
})
