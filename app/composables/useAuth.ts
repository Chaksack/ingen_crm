export interface AuthUser {
  id: string
  email: string
  name: string
  avatar?: string | null
  role: 'admin' | 'manager' | 'staff'
}

export function useAuth() {
  const user = useState<AuthUser | null>('auth-user', () => null)
  const initialized = useState('auth-initialized', () => false)

  async function fetchUser() {
    try {
      user.value = await $fetch<AuthUser>('/api/auth/me')
    }
    catch {
      user.value = null
    }
    finally {
      initialized.value = true
    }
  }

  async function login(email: string, password: string) {
    await $fetch('/api/auth/login', { method: 'POST', body: { email, password } })
  }

  async function verifyOtp(code: string) {
    return $fetch<{ purpose: 'login' | 'password_reset' }>('/api/auth/verify-otp', {
      method: 'POST',
      body: { code },
    })
  }

  async function resendOtp() {
    await $fetch('/api/auth/resend-otp', { method: 'POST' })
  }

  async function forgotPassword(email: string) {
    await $fetch('/api/auth/forgot-password', { method: 'POST', body: { email } })
  }

  async function resetPassword(password: string) {
    await $fetch('/api/auth/reset-password', { method: 'POST', body: { password } })
  }

  async function logout() {
    await $fetch('/api/auth/logout', { method: 'POST' })
    user.value = null
    await navigateTo('/')
  }

  return { user, initialized, fetchUser, login, verifyOtp, resendOtp, forgotPassword, resetPassword, logout }
}
