import { desc } from 'drizzle-orm'
import { db } from '../../db/client'
import { clients } from '../../db/schema'

export default defineEventHandler(async (event) => {
  await requireUser(event)
  return db.select().from(clients).orderBy(desc(clients.createdAt))
})
