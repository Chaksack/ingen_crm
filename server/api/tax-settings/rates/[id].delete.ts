import { eq } from 'drizzle-orm'
import { db } from '../../../db/client'
import { taxRates } from '../../../db/schema'

export default defineEventHandler(async (event) => {
  await requireUser(event)
  const id = getRouterParam(event, 'id')!
  await db.delete(taxRates).where(eq(taxRates.id, id))
  return { ok: true }
})
