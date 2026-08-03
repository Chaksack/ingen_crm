import { desc } from 'drizzle-orm'
import { db } from '../../db/client'
import { expenses } from '../../db/schema'

export default defineEventHandler(async (event) => {
  await requireUser(event)
  return db.query.expenses.findMany({
    orderBy: desc(expenses.expenseDate),
    with: { vendor: true },
  })
})
