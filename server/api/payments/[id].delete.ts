import { eq } from 'drizzle-orm'
import { db } from '../../db/client'
import { payments } from '../../db/schema'

export default defineEventHandler(async (event) => {
  await requireUser(event)
  const id = getRouterParam(event, 'id')!
  await db.delete(payments).where(eq(payments.id, id))
  return { ok: true }
})
