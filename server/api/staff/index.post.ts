import { eq } from 'drizzle-orm'
import { z } from 'zod'
import { db } from '../../db/client'
import { authCodes, staff, users } from '../../db/schema'

const bodySchema = z.object({
  firstName: z.string().min(1),
  lastName: z.string().min(1),
  email: z.string().email(),
  phoneNumber: z.string().optional(),
  role: z.enum(['Staff', 'Manager', 'Admin']).optional(),
  department: z.string().optional(),
  status: z.enum(['active', 'inactive']).optional(),
  avatar: z.string().optional(),
  createLogin: z.boolean().optional(),
})

export default defineEventHandler(async (event) => {
  const session = await requireUser(event)
  const body = await readValidatedBody(event, bodySchema.parse)
  const { createLogin, ...staffFields } = body

  if (createLogin && session.role !== 'admin') {
    throw createError({ statusCode: 403, statusMessage: 'Only an admin can grant system access' })
  }

  if (createLogin) {
    const existingUser = await db.query.users.findFirst({ where: eq(users.email, body.email) })
    if (existingUser) {
      throw createError({ statusCode: 409, statusMessage: 'A login already exists for this email' })
    }
  }

  return db.transaction(async (tx) => {
    const [staffRow] = await tx.insert(staff).values(staffFields).returning()

    if (createLogin) {
      const [user] = await tx.insert(users).values({
        email: body.email,
        name: `${body.firstName} ${body.lastName}`,
        role: (body.role ?? 'Staff').toLowerCase() as 'admin' | 'manager' | 'staff',
        status: 'pending',
        staffId: staffRow.id,
      }).returning()

      const token = generateToken()
      await tx.insert(authCodes).values({
        userId: user.id,
        code: token,
        purpose: 'invite',
        expiresAt: new Date(Date.now() + 7 * 24 * 60 * 60 * 1000),
      })

      const link = `${getRequestURL(event).origin}/accept-invite?token=${token}`
      await sendStaffInviteEmail(user.email, user.name, link)
    }

    return staffRow
  })
})
