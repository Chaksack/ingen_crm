import { eq } from 'drizzle-orm'
import { db } from '../../db/client'
import { invoices } from '../../db/schema'

export default defineEventHandler(async (event) => {
  await requireUser(event)
  const id = getRouterParam(event, 'id')!
  await db.delete(invoices).where(eq(invoices.id, id))
  return { ok: true }
})
