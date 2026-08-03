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

// Serverless functions (Vercel, etc.) spin up fresh instances per invocation, so each
// one should hold at most a couple of connections rather than postgres.js's default pool of 10.
const queryClient = postgres(connectionString, { max: 3, idle_timeout: 20, connect_timeout: 10 })

export const db = drizzle(queryClient, { schema, casing: 'snake_case' })
