import { eq } from 'drizzle-orm'
import { z } from 'zod'
import { db } from '../../../db/client'
import { clients, supportTickets } from '../../../db/schema'

const bodySchema = z.object({
  name: z.string().min(1),
  email: z.string().email(),
  phone: z.string().optional(),
  company: z.string().optional(),
  subject: z.string().min(1),
  description: z.string().min(1),
  category: z.string().optional(),
  priority: z.enum(['low', 'medium', 'high', 'urgent']).optional(),
})

export default defineEventHandler(async (event) => {
  const body = await readValidatedBody(event, bodySchema.parse)

  const matchedClient = await db.query.clients.findFirst({ where: eq(clients.email, body.email) })
  const ticketNumber = await nextDocumentNumber('TKT')

  const [ticket] = await db.insert(supportTickets).values({
    ticketNumber,
    name: body.name,
    email: body.email,
    phone: body.phone,
    company: body.company,
    subject: body.subject,
    description: body.description,
    category: body.category,
    priority: body.priority ?? 'medium',
    clientId: matchedClient?.id,
  }).returning()

  return { ticketNumber: ticket.ticketNumber, id: ticket.id }
})
