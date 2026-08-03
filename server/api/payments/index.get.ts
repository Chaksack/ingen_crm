import { desc, eq } from 'drizzle-orm'
import { db } from '../../db/client'
import { payments } from '../../db/schema'

export default defineEventHandler(async (event) => {
  await requireUser(event)
  const query = getQuery(event)
  const invoiceId = typeof query.invoiceId === 'string' ? query.invoiceId : undefined

  return db.query.payments.findMany({
    where: invoiceId ? eq(payments.invoiceId, invoiceId) : undefined,
    orderBy: desc(payments.createdAt),
    with: { invoice: { with: { client: true } } },
  })
})
