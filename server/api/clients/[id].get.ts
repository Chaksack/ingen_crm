import { eq } from 'drizzle-orm'
import { db } from '../../db/client'
import { clientBankAccounts, clientMomoAccounts, clients } from '../../db/schema'

export default defineEventHandler(async (event) => {
  await requireUser(event)
  const id = getRouterParam(event, 'id')!

  const client = await db.query.clients.findFirst({ where: eq(clients.id, id) })
  if (!client) {
    throw createError({ statusCode: 404, statusMessage: 'Client not found' })
  }

  const [bankAccounts, momoAccounts] = await Promise.all([
    db.select().from(clientBankAccounts).where(eq(clientBankAccounts.clientId, id)),
    db.select().from(clientMomoAccounts).where(eq(clientMomoAccounts.clientId, id)),
  ])

  return { ...client, bankAccounts, momoAccounts }
})
