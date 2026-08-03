import { desc } from 'drizzle-orm'
import { db } from '../../db/client'
import { financialInstitutions } from '../../db/schema'

export default defineEventHandler(async (event) => {
  await requireUser(event)
  return db.select().from(financialInstitutions).orderBy(desc(financialInstitutions.createdAt))
})
