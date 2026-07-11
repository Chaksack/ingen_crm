// Redirect away to "/" if the visitor is already authenticated.
const publicOnlyPages = ['/login', '/register']
// Reachable regardless of auth state (e.g. an invite link opened while
// already logged in to a different org).
const alwaysPublicPages = ['/accept-invite']

export default defineNuxtRouteMiddleware(async (to) => {
  const auth = useAuthStore()

  if (!auth.token) {
    auth.hydrate()
  }

  if (alwaysPublicPages.includes(to.path)) return

  const isPublic = publicOnlyPages.includes(to.path)

  if (!auth.token) {
    if (!isPublic) return navigateTo('/login')
    return
  }

  if (!auth.user) {
    try {
      const api = useApi()
      auth.setUser(await api('/me'))
      auth.setManifest(await api('/me/manifest'))
    } catch {
      auth.logout()
      return
    }
  }

  if (isPublic) return navigateTo('/')
})
