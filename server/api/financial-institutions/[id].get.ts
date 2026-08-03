import { eq } from 'drizzle-orm'
import { db } from '../../db/client'
import { financialInstitutions } from '../../db/schema'

export default defineEventHandler(async (event) => {
  await requireUser(event)
  const id = getRouterParam(event, 'id')!
  const row = await db.query.financialInstitutions.findFirst({ where: eq(financialInstitutions.id, id) })
  if (!row) {
    throw createError({ statusCode: 404, statusMessage: 'Financial institution not found' })
  }
  return row
})
