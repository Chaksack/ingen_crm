import { asc } from 'drizzle-orm'
import { db } from '../../db/client'
import { taxRates } from '../../db/schema'

export default defineEventHandler(async (event) => {
  await requireUser(event)

  const [settings, rates] = await Promise.all([
    db.query.companyTaxSettings.findFirst(),
    db.select().from(taxRates).orderBy(asc(taxRates.order)),
  ])

  return { settings: settings ?? null, rates }
})
