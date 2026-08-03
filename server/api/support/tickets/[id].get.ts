import { asc, eq } from 'drizzle-orm'
import { db } from '../../../db/client'
import { supportTicketMessages, supportTickets } from '../../../db/schema'

export default defineEventHandler(async (event) => {
  await requireUser(event)
  const id = getRouterParam(event, 'id')!

  const ticket = await db.query.supportTickets.findFirst({
    where: eq(supportTickets.id, id),
    with: {
      client: true,
      assignee: { columns: { id: true, name: true, email: true, avatar: true } },
    },
  })
  if (!ticket) {
    throw createError({ statusCode: 404, statusMessage: 'Ticket not found' })
  }

  const messages = await db.select().from(supportTicketMessages).where(eq(supportTicketMessages.ticketId, id)).orderBy(asc(supportTicketMessages.createdAt))

  return { ...ticket, messages }
})
