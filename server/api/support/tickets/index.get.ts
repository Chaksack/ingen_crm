import { desc } from 'drizzle-orm'
import { db } from '../../../db/client'
import { supportTickets } from '../../../db/schema'

export default defineEventHandler(async (event) => {
  await requireUser(event)
  return db.query.supportTickets.findMany({
    orderBy: desc(supportTickets.createdAt),
    with: {
      client: true,
      assignee: { columns: { id: true, name: true, email: true, avatar: true } },
    },
  })
})
