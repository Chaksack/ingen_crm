import { eq } from 'drizzle-orm'
import { db } from '../../db/client'
import { expenses } from '../../db/schema'

export default defineEventHandler(async (event) => {
  await requireUser(event)
  const id = getRouterParam(event, 'id')!
  const row = await db.query.expenses.findFirst({ where: eq(expenses.id, id) })
  if (!row) {
    throw createError({ statusCode: 404, statusMessage: 'Expense not found' })
  }
  return row
})
