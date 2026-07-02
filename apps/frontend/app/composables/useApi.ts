export function useApi() {
  const config = useRuntimeConfig()
  const auth = useAuthStore()

  return $fetch.create({
    baseURL: `${config.public.apiBase}/api/v1`,
    onRequest({ options }) {
      if (auth.token) {
        options.headers = new Headers(options.headers)
        options.headers.set('Authorization', `Bearer ${auth.token}`)
      }
    },
    onResponseError({ response }) {
      if (response.status === 401) {
        auth.logout()
      }
    },
  })
}
