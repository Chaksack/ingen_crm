import { randomBytes } from 'node:crypto'
import process from 'node:process'
import bcrypt from 'bcryptjs'
import { eq } from 'drizzle-orm'
import { db } from './client'
import { companyTaxSettings, taxRates, users } from './schema'

async function seedAdmin() {
  const email = 'chakdahah@gmail.com'
  const existing = await db.query.users.findFirst({ where: eq(users.email, email) })
  if (existing) {
    console.warn(`Admin user already exists: ${email}`)
    return
  }

  const password = randomBytes(9).toString('base64url')
  const passwordHash = await bcrypt.hash(password, 12)

  await db.insert(users).values({
    email,
    passwordHash,
    name: 'Andrew Chakdahah',
    role: 'admin',
    status: 'active',
  })

  console.warn('Seeded default admin user:')
  console.warn(`  email:    ${email}`)
  console.warn(`  password: ${password}`)
  console.warn('Change this password after first login.')
}

async function seedTaxDefaults() {
  const existingRates = await db.select().from(taxRates).limit(1)
  if (existingRates.length === 0) {
    await db.insert(taxRates).values([
      { name: 'NHIL', rate: '2.5', compound: false, order: 0, active: true },
      { name: 'GETFund', rate: '2.5', compound: false, order: 1, active: true },
      { name: 'COVID-19 Levy', rate: '1', compound: false, order: 2, active: true },
      { name: 'VAT', rate: '15', compound: true, order: 3, active: true },
    ])
    console.warn('Seeded default GRA tax rates: NHIL 2.5%, GETFund 2.5%, COVID-19 Levy 1%, VAT 15% (compound)')
  }
  else {
    console.warn('Tax rates already seeded, skipping.')
  }

  const existingSettings = await db.query.companyTaxSettings.findFirst()
  if (!existingSettings) {
    await db.insert(companyTaxSettings).values({
      companyName: 'Ingenicx',
    })
    console.warn('Seeded blank company tax settings row (set TIN/VAT number under Finance > Tax Settings).')
  }
  else {
    console.warn('Company tax settings already exist, skipping.')
  }
}

async function main() {
  await seedAdmin()
  await seedTaxDefaults()
}

main()
  .then(() => process.exit(0))
  .catch((err) => {
    console.error(err)
    process.exit(1)
  })
