import { randomUUID } from 'node:crypto'
import { mkdir, writeFile } from 'node:fs/promises'
import { extname, join } from 'node:path'
import process from 'node:process'

const UPLOAD_DIR = join(process.cwd(), 'public', 'uploads', 'receipts')

export default defineEventHandler(async (event) => {
  await requireUser(event)

  const parts = await readMultipartFormData(event)
  const file = parts?.find(p => p.name === 'file')
  if (!file || !file.filename) {
    throw createError({ statusCode: 400, statusMessage: 'No file uploaded' })
  }

  await mkdir(UPLOAD_DIR, { recursive: true })
  const filename = `${randomUUID()}${extname(file.filename)}`
  await writeFile(join(UPLOAD_DIR, filename), file.data)

  return { url: `/uploads/receipts/${filename}` }
})
