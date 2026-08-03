import { eq } from 'drizzle-orm'
import { db } from '../../db/client'
import { vendors } from '../../db/schema'

export default defineEventHandler(async (event) => {
  await requireUser(event)
  const id = getRouterParam(event, 'id')!
  const row = await db.query.vendors.findFirst({ where: eq(vendors.id, id) })
  if (!row) {
    throw createError({ statusCode: 404, statusMessage: 'Vendor not found' })
  }
  return row
})
