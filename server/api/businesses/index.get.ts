import { desc } from 'drizzle-orm'
import { db } from '../../db/client'
import { businesses } from '../../db/schema'

export default defineEventHandler(async (event) => {
  await requireUser(event)
  return db.select().from(businesses).orderBy(desc(businesses.createdAt))
})
