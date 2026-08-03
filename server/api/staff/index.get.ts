import { desc } from 'drizzle-orm'
import { db } from '../../db/client'
import { staff } from '../../db/schema'

export default defineEventHandler(async (event) => {
  await requireUser(event)
  return db.select().from(staff).orderBy(desc(staff.createdAt))
})
