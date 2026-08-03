const GUEST_PAGES = new Set(['/', '/forgot-password', '/otp', '/new-password'])
const PUBLIC_PAGES = new Set(['/401', '/403', '/404', '/500', '/503', '/support', '/accept-invite'])

export default defineNuxtRouteMiddleware(async (to) => {
  if (PUBLIC_PAGES.has(to.path))
    return

  const { user } = useAuth()

  if (import.meta.server || !user.value) {
    const fetcher = import.meta.server ? useRequestFetch() : $fetch
    try {
      user.value = await fetcher('/api/auth/me') as any
    }
    catch {
      user.value = null
    }
  }

  if (!user.value && !GUEST_PAGES.has(to.path))
    return navigateTo('/')

  if (user.value && GUEST_PAGES.has(to.path))
    return navigateTo('/dashboard')
})
