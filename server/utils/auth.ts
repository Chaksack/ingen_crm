import type { H3Event } from 'h3'
import process from 'node:process'
import bcrypt from 'bcryptjs'
import { jwtVerify, SignJWT } from 'jose'

const SESSION_COOKIE = 'session'
const PENDING_AUTH_COOKIE = 'pending_auth'
const RESET_VERIFIED_COOKIE = 'reset_verified'

function secretKey() {
  const secret = process.env.JWT_SECRET
  if (!secret)
    throw new Error('JWT_SECRET is not set')
  return new TextEncoder().encode(secret)
}

export interface SessionPayload {
  sub: string
  role: 'admin' | 'manager' | 'staff'
}

export interface PendingAuthPayload {
  sub: string
  purpose: 'login' | 'reset'
}

export function hashPassword(password: string) {
  return bcrypt.hash(password, 12)
}

export function verifyPassword(password: string, hash: string) {
  return bcrypt.compare(password, hash)
}

async function sign(payload: Record<string, unknown>, expiresIn: string) {
  return new SignJWT(payload)
    .setProtectedHeader({ alg: 'HS256' })
    .setIssuedAt()
    .setExpirationTime(expiresIn)
    .sign(secretKey())
}

async function verify<T>(token: string): Promise<T | null> {
  try {
    const { payload } = await jwtVerify(token, secretKey())
    return payload as T
  }
  catch {
    return null
  }
}

const isProd = process.env.NODE_ENV === 'production'

export async function setSessionCookie(event: H3Event, user: { id: string, role: string }) {
  const token = await sign({ sub: user.id, role: user.role }, '7d')
  setCookie(event, SESSION_COOKIE, token, {
    httpOnly: true,
    secure: isProd,
    sameSite: 'lax',
    path: '/',
    maxAge: 60 * 60 * 24 * 7,
  })
}

export function clearSessionCookie(event: H3Event) {
  deleteCookie(event, SESSION_COOKIE, { path: '/' })
}

export async function getSessionUser(event: H3Event): Promise<SessionPayload | null> {
  const token = getCookie(event, SESSION_COOKIE)
  if (!token)
    return null
  return verify<SessionPayload>(token)
}

export async function requireUser(event: H3Event): Promise<SessionPayload> {
  const user = await getSessionUser(event)
  if (!user) {
    throw createError({ statusCode: 401, statusMessage: 'Unauthorized' })
  }
  return user
}

export async function setPendingAuthCookie(event: H3Event, userId: string, purpose: 'login' | 'reset') {
  const token = await sign({ sub: userId, purpose }, '10m')
  setCookie(event, PENDING_AUTH_COOKIE, token, {
    httpOnly: true,
    secure: isProd,
    sameSite: 'lax',
    path: '/',
    maxAge: 60 * 10,
  })
}

export async function getPendingAuthCookie(event: H3Event): Promise<PendingAuthPayload | null> {
  const token = getCookie(event, PENDING_AUTH_COOKIE)
  if (!token)
    return null
  return verify<PendingAuthPayload>(token)
}

export function clearPendingAuthCookie(event: H3Event) {
  deleteCookie(event, PENDING_AUTH_COOKIE, { path: '/' })
}

export async function setResetVerifiedCookie(event: H3Event, userId: string) {
  const token = await sign({ sub: userId }, '10m')
  setCookie(event, RESET_VERIFIED_COOKIE, token, {
    httpOnly: true,
    secure: isProd,
    sameSite: 'lax',
    path: '/',
    maxAge: 60 * 10,
  })
}

export async function getResetVerifiedCookie(event: H3Event): Promise<{ sub: string } | null> {
  const token = getCookie(event, RESET_VERIFIED_COOKIE)
  if (!token)
    return null
  return verify<{ sub: string }>(token)
}

export function clearResetVerifiedCookie(event: H3Event) {
  deleteCookie(event, RESET_VERIFIED_COOKIE, { path: '/' })
}
