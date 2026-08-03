import { eq } from 'drizzle-orm'
import { db } from '../../db/client'
import { businessBankAccounts, businesses } from '../../db/schema'

export default defineEventHandler(async (event) => {
  await requireUser(event)
  const id = getRouterParam(event, 'id')!

  const business = await db.query.businesses.findFirst({ where: eq(businesses.id, id) })
  if (!business) {
    throw createError({ statusCode: 404, statusMessage: 'Business not found' })
  }

  const bankAccounts = await db.select().from(businessBankAccounts).where(eq(businessBankAccounts.businessId, id))

  return { ...business, bankAccounts }
})
