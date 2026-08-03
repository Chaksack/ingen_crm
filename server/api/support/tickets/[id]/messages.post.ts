import { eq } from 'drizzle-orm'
import { z } from 'zod'
import { db } from '../../../../db/client'
import { supportTicketMessages, supportTickets, users } from '../../../../db/schema'

const bodySchema = z.object({
  body: z.string().min(1),
})

export default defineEventHandler(async (event) => {
  const session = await requireUser(event)
  const ticketId = getRouterParam(event, 'id')!
  const body = await readValidatedBody(event, bodySchema.parse)

  const ticket = await db.query.supportTickets.findFirst({ where: eq(supportTickets.id, ticketId) })
  if (!ticket) {
    throw createError({ statusCode: 404, statusMessage: 'Ticket not found' })
  }

  const author = await db.query.users.findFirst({ where: eq(users.id, session.sub) })

  const [message] = await db.insert(supportTicketMessages).values({
    ticketId,
    authorType: 'staff',
    authorName: author?.name ?? 'Staff',
    body: body.body,
  }).returning()

  if (ticket.status === 'open') {
    await db.update(supportTickets).set({ status: 'in_progress' }).where(eq(supportTickets.id, ticketId))
  }

  return message
})
