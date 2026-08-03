import { asc, eq } from 'drizzle-orm'
import { db } from '../../db/client'
import { quotationLineItems, quotations, quotationTaxes } from '../../db/schema'

export default defineEventHandler(async (event) => {
  await requireUser(event)
  const id = getRouterParam(event, 'id')!

  const quotation = await db.query.quotations.findFirst({
    where: eq(quotations.id, id),
    with: { client: true },
  })
  if (!quotation) {
    throw createError({ statusCode: 404, statusMessage: 'Quotation not found' })
  }

  const [lineItems, taxes] = await Promise.all([
    db.select().from(quotationLineItems).where(eq(quotationLineItems.quotationId, id)).orderBy(asc(quotationLineItems.order)),
    db.select().from(quotationTaxes).where(eq(quotationTaxes.quotationId, id)).orderBy(asc(quotationTaxes.order)),
  ])

  return { ...quotation, lineItems, taxes }
})
