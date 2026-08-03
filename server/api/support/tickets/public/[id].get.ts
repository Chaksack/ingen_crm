import { and, asc, eq } from 'drizzle-orm'
import { db } from '../../../../db/client'
import { supportTicketMessages, supportTickets } from '../../../../db/schema'

export default defineEventHandler(async (event) => {
  const id = getRouterParam(event, 'id')!
  const { token } = getQuery(event)

  if (!token || typeof token !== 'string') {
    throw createError({ statusCode: 401, statusMessage: 'Missing access token' })
  }

  const ticket = await db.query.supportTickets.findFirst({
    where: and(eq(supportTickets.id, id), eq(supportTickets.accessToken, token)),
    columns: {
      id: true,
      ticketNumber: true,
      name: true,
      subject: true,
      status: true,
      preferredContact: true,
      createdAt: true,
    },
  })
  // 404 rather than 401/403 for a token mismatch too, so a guessed id can't be used to
  // probe whether a ticket exists.
  if (!ticket) {
    throw createError({ statusCode: 404, statusMessage: 'Ticket not found' })
  }

  const messages = await db.select().from(supportTicketMessages).where(eq(supportTicketMessages.ticketId, id)).orderBy(asc(supportTicketMessages.createdAt))

  return { ...ticket, messages }
})
