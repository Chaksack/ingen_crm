import { and, eq } from 'drizzle-orm'
import { z } from 'zod'
import { db } from '../../../../../db/client'
import { supportTicketMessages, supportTickets } from '../../../../../db/schema'

const bodySchema = z.object({
  token: z.string().min(1),
  body: z.string().min(1),
})

export default defineEventHandler(async (event) => {
  const id = getRouterParam(event, 'id')!
  const { token, body } = await readValidatedBody(event, bodySchema.parse)

  const ticket = await db.query.supportTickets.findFirst({
    where: and(eq(supportTickets.id, id), eq(supportTickets.accessToken, token)),
  })
  if (!ticket) {
    throw createError({ statusCode: 404, statusMessage: 'Ticket not found' })
  }

  const [message] = await db.insert(supportTicketMessages).values({
    ticketId: id,
    authorType: 'client',
    authorName: ticket.name,
    body,
  }).returning()

  // A reply from the customer means the ticket needs attention again.
  if (ticket.status === 'resolved' || ticket.status === 'closed') {
    await db.update(supportTickets).set({ status: 'open' }).where(eq(supportTickets.id, id))
  }

  return message
})
