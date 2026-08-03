import { z } from 'zod'
import { db } from '../../db/client'
import { businessBankAccounts, businesses } from '../../db/schema'

const bodySchema = z.object({
  companyName: z.string().min(1),
  companyNumber: z.string().optional(),
  email: z.string().email(),
  phoneNumber: z.string().optional(),
  address: z.string().optional(),
  avatar: z.string().optional(),
  capital: z.union([z.number(), z.string()]).optional(),
  foundedYear: z.number().optional(),
  country: z.string().optional(),
  currency: z.string().optional(),
  creditScore: z.number().optional(),
  kyc: z.boolean().optional(),
  status: z.enum(['active', 'inactive', 'pending']).optional(),
  bankAccounts: z.array(z.object({
    bankName: z.string().min(1),
    accountNumber: z.string().min(1),
    accountHolder: z.string().min(1),
  })).optional(),
})

export default defineEventHandler(async (event) => {
  await requireUser(event)
  const body = await readValidatedBody(event, bodySchema.parse)
  const { bankAccounts, ...fields } = body

  const [business] = await db.insert(businesses).values({
    ...fields,
    capital: fields.capital !== undefined ? String(fields.capital) : undefined,
  }).returning()

  if (bankAccounts?.length) {
    await db.insert(businessBankAccounts).values(bankAccounts.map(a => ({ ...a, businessId: business.id })))
  }

  return business
})
