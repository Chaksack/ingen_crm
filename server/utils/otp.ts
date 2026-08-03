import { randomBytes, randomInt } from 'node:crypto'
import process from 'node:process'
import { Resend } from 'resend'

export function generateCode() {
  return String(randomInt(0, 1_000_000)).padStart(6, '0')
}

export function generateToken() {
  return randomBytes(24).toString('hex')
}

const isProd = process.env.NODE_ENV === 'production'

function getResend() {
  const apiKey = process.env.RESEND_API_KEY
  if (!apiKey)
    return null
  return new Resend(apiKey)
}

async function sendEmail(to: string, subject: string, html: string, devLabel: string, devValue: string) {
  const resend = getResend()
  if (!resend) {
    if (!isProd)
      console.warn(`[dev] ${devLabel} for ${to}: ${devValue}`)
    return
  }

  try {
    await resend.emails.send({
      from: 'Ingenicx <onboarding@resend.dev>',
      to,
      subject,
      html,
    })
  }
  finally {
    if (!isProd)
      console.warn(`[dev] ${devLabel} for ${to}: ${devValue}`)
  }
}

export async function sendLoginOtpEmail(to: string, code: string) {
  await sendEmail(
    to,
    'Your Ingenicx sign-in code',
    `<p>Your sign-in verification code is:</p><h2>${code}</h2><p>This code expires in 10 minutes.</p>`,
    'Login OTP code',
    code,
  )
}

export async function sendPasswordResetEmail(to: string, code: string) {
  await sendEmail(
    to,
    'Reset your Ingenicx password',
    `<p>Your password reset code is:</p><h2>${code}</h2><p>This code expires in 10 minutes.</p>`,
    'Password reset code',
    code,
  )
}

export async function sendStaffInviteEmail(to: string, name: string, link: string) {
  await sendEmail(
    to,
    'You\'ve been invited to Ingenicx',
    `<p>Hi ${name},</p><p>You've been added as a staff member on Ingenicx. Click the link below to set your password and activate your account:</p><p><a href="${link}">${link}</a></p><p>This link expires in 7 days.</p>`,
    'Staff invite link',
    link,
  )
}
