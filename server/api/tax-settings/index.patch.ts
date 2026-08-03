import { eq } from 'drizzle-orm'
import { z } from 'zod'
import { db } from '../../db/client'
import { companyTaxSettings } from '../../db/schema'

const bodySchema = z.object({
  companyName: z.string().min(1).optional(),
  tin: z.string().optional(),
  vatNumber: z.string().optional(),
  address: z.string().optional(),
  phone: z.string().optional(),
  email: z.string().optional(),
})

export default defineEventHandler(async (event) => {
  await requireUser(event)
  const body = await readValidatedBody(event, bodySchema.parse)

  const existing = await db.query.companyTaxSettings.findFirst()

  if (!existing) {
    const [created] = await db.insert(companyTaxSettings).values(body).returning()
    return created
  }

  const [updated] = await db.update(companyTaxSettings)
    .set(body)
    .where(eq(companyTaxSettings.id, existing.id))
    .returning()

  return updated
})
