import { desc } from 'drizzle-orm'
import { db } from '../../db/client'
import { vendors } from '../../db/schema'

export default defineEventHandler(async (event) => {
  await requireUser(event)
  return db.select().from(vendors).orderBy(desc(vendors.createdAt))
})
