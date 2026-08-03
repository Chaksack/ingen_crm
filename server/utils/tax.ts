import { asc, eq } from 'drizzle-orm'
import { db } from '../db/client'
import { taxRates } from '../db/schema'

export interface TaxRateDef {
  name: string
  rate: number
  compound: boolean
}

export interface TaxLine {
  name: string
  rate: number
  amount: number
  compound: boolean
  order: number
}

function round2(n: number) {
  return Math.round(n * 100) / 100
}

export async function getActiveTaxRates(): Promise<TaxRateDef[]> {
  const rows = await db.select().from(taxRates).where(eq(taxRates.active, true)).orderBy(asc(taxRates.order))
  return rows.map(r => ({ name: r.name, rate: Number(r.rate), compound: r.compound }))
}

/**
 * GRA-style cascading tax calculation: non-compound levies (e.g. NHIL, GETFund,
 * COVID Levy) are calculated on the taxable base; compound taxes (e.g. VAT) are
 * then calculated on the taxable base plus those levies.
 */
export function computeTaxLines(taxableBase: number, rates: TaxRateDef[]): { lines: TaxLine[], taxAmount: number } {
  const nonCompound = rates.filter(r => !r.compound)
  const compound = rates.filter(r => r.compound)

  const lines: TaxLine[] = []
  let order = 0
  let leviesTotal = 0

  for (const r of nonCompound) {
    const amount = round2(taxableBase * (r.rate / 100))
    leviesTotal += amount
    lines.push({ name: r.name, rate: r.rate, amount, compound: false, order: order++ })
  }

  const compoundBase = taxableBase + leviesTotal
  for (const r of compound) {
    const amount = round2(compoundBase * (r.rate / 100))
    lines.push({ name: r.name, rate: r.rate, amount, compound: true, order: order++ })
  }

  const taxAmount = round2(lines.reduce((sum, l) => sum + l.amount, 0))
  return { lines, taxAmount }
}
