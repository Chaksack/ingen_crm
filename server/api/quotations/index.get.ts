import { desc } from 'drizzle-orm'
import { db } from '../../db/client'
import { quotations } from '../../db/schema'

export default defineEventHandler(async (event) => {
  await requireUser(event)
  return db.query.quotations.findMany({
    orderBy: desc(quotations.createdAt),
    with: { client: true },
  })
})
