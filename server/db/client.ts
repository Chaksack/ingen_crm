import process from 'node:process'
import { config } from 'dotenv'
import { drizzle } from 'drizzle-orm/postgres-js'
import postgres from 'postgres'
import * as schema from './schema'

// Nitro injects .env vars automatically at runtime; standalone scripts (seed, etc.)
// running outside Nitro need it loaded explicitly, regardless of their import order.
if (!process.env.DATABASE_URL) {
  config()
}

const connectionString = process.env.DATABASE_URL
if (!connectionString)
  throw new Error('DATABASE_URL is not set')

const queryClient = postgres(connectionString)

export const db = drizzle(queryClient, { schema, casing: 'snake_case' })
