import { defineStore } from 'pinia'

export interface ManifestNavItem {
  key: string
  label: string
  path: string
  icon: string
}

export interface Manifest {
  modules: string[]
  nav: ManifestNavItem[]
}

export interface CurrentUser {
  id: string
  email: string
  display_name: string
  organization_id: string
  business_unit_id: string | null
  roles: string[]
}

export const useAuthStore = defineStore('auth', {
  state: () => ({
    token: null as string | null,
    user: null as CurrentUser | null,
    manifest: null as Manifest | null,
  }),
  getters: {
    isAuthenticated: (state) => !!state.token,
  },
  actions: {
    hydrate() {
      if (import.meta.client) {
        this.token = localStorage.getItem('ingen_token')
      }
    },
    setToken(token: string) {
      this.token = token
      if (import.meta.client) {
        localStorage.setItem('ingen_token', token)
      }
    },
    setUser(user: CurrentUser) {
      this.user = user
    },
    setManifest(manifest: Manifest) {
      this.manifest = manifest
    },
    logout() {
      this.token = null
      this.user = null
      this.manifest = null
      if (import.meta.client) {
        localStorage.removeItem('ingen_token')
      }
      navigateTo('/login')
    },
  },
})
