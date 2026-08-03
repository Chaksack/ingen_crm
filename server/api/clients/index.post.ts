import { z } from 'zod'
import { db } from '../../db/client'
import { clientBankAccounts, clientMomoAccounts, clients } from '../../db/schema'

const bodySchema = z.object({
  name: z.string().min(1),
  contactPerson: z.string().optional(),
  email: z.string().email(),
  phone: z.string().optional(),
  address: z.string().optional(),
  notes: z.string().optional(),
  status: z.enum(['active', 'inactive']).optional(),
  creditScore: z.number().optional(),
  kyc: z.boolean().optional(),
  dob: z.string().optional(),
  nationality: z.string().optional(),
  idType: z.string().optional(),
  idNumber: z.string().optional(),
  monthlyIncome: z.union([z.number(), z.string()]).optional(),
  avatar: z.string().optional(),
  bankAccounts: z.array(z.object({
    bankName: z.string().min(1),
    accountNumber: z.string().min(1),
    accountHolder: z.string().min(1),
  })).optional(),
  momoAccounts: z.array(z.object({
    networkName: z.string().min(1),
    momoNumber: z.string().min(1),
    accountHolder: z.string().min(1),
  })).optional(),
})

export default defineEventHandler(async (event) => {
  await requireUser(event)
  const body = await readValidatedBody(event, bodySchema.parse)
  const { bankAccounts, momoAccounts, ...clientFields } = body

  const [client] = await db.insert(clients).values({
    ...clientFields,
    monthlyIncome: clientFields.monthlyIncome !== undefined ? String(clientFields.monthlyIncome) : undefined,
  }).returning()

  if (bankAccounts?.length) {
    await db.insert(clientBankAccounts).values(bankAccounts.map(a => ({ ...a, clientId: client.id })))
  }
  if (momoAccounts?.length) {
    await db.insert(clientMomoAccounts).values(momoAccounts.map(a => ({ ...a, clientId: client.id })))
  }

  return client
})
