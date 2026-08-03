import { sql } from 'drizzle-orm'
import { db } from '../db/client'
import { documentSequences } from '../db/schema'

export async function nextDocumentNumber(prefix: 'INV' | 'QUO' | 'TKT' | 'RCT') {
  const year = new Date().getFullYear()
  const key = `${prefix}-${year}`

  const rows = await db
    .insert(documentSequences)
    .values({ key, value: 1 })
    .onConflictDoUpdate({
      target: documentSequences.key,
      set: { value: sql`${documentSequences.value} + 1` },
    })
    .returning({ value: documentSequences.value })

  const seq = rows[0].value
  return `${key}-${String(seq).padStart(4, '0')}`
}
