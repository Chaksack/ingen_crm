import { eq } from 'drizzle-orm'
import { db } from '../../db/client'
import { staff } from '../../db/schema'

export default defineEventHandler(async (event) => {
  await requireUser(event)
  const id = getRouterParam(event, 'id')!
  const row = await db.query.staff.findFirst({ where: eq(staff.id, id) })
  if (!row) {
    throw createError({ statusCode: 404, statusMessage: 'Staff not found' })
  }
  return row
})
