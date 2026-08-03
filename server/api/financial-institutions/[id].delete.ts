import { eq } from 'drizzle-orm'
import { db } from '../../db/client'
import { financialInstitutions } from '../../db/schema'

export default defineEventHandler(async (event) => {
  await requireUser(event)
  const id = getRouterParam(event, 'id')!
  await db.delete(financialInstitutions).where(eq(financialInstitutions.id, id))
  return { ok: true }
})
