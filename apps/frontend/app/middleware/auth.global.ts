const publicPages = ['/login', '/register']

export default defineNuxtRouteMiddleware(async (to) => {
  const auth = useAuthStore()

  if (!auth.token) {
    auth.hydrate()
  }

  const isPublic = publicPages.includes(to.path)

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
