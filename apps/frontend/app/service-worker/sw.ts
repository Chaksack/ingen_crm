/// <reference lib="webworker" />
import { precacheAndRoute } from 'workbox-precaching'

declare let self: ServiceWorkerGlobalScope

// Injected at build time with the app shell's asset list — see
// nuxt.config.ts `pwa.injectManifest`. Equivalent to generateSW's
// precaching, but injectManifest is required to also handle push events
// below (generateSW's fixed template has no room for custom listeners).
//
// No navigation-fallback route: this app serves from a Nitro node-server
// (not a static `nuxt generate` build), so there's no static index.html
// asset to precache and hand back for an offline reload — a
// createHandlerBoundToURL('/')-style route throws at SW-script evaluation
// time if '/' was never actually precached, which would break registration
// entirely. Offline full-reload support would need a runtime cache-first
// route for navigations instead; out of scope for now (see README).
precacheAndRoute(self.__WB_MANIFEST)

// Web Push (PRD §5.11/§7.5): the payload shape is set in
// apps/api/internal/push/push.go — {title, body, url}.
self.addEventListener('push', (event) => {
  if (!event.data) return
  let data: { title?: string; body?: string; url?: string }
  try {
    data = event.data.json()
  } catch {
    return
  }
  const title = data.title || 'IngenCore'
  event.waitUntil(
    self.registration.showNotification(title, {
      body: data.body || '',
      icon: '/pwa-192x192.png',
      badge: '/pwa-192x192.png',
      data: { url: data.url || '/' },
    }),
  )
})

// Tapping a notification deep-links into the record/chat it's about,
// focusing an existing tab if one is already open.
self.addEventListener('notificationclick', (event) => {
  event.notification.close()
  const url = (event.notification.data as { url?: string } | undefined)?.url || '/'
  event.waitUntil(
    self.clients.matchAll({ type: 'window', includeUncontrolled: true }).then((clientList) => {
      for (const client of clientList) {
        if ('focus' in client && 'navigate' in client) {
          void (client as WindowClient).navigate(url)
          return client.focus()
        }
      }
      return self.clients.openWindow(url)
    }),
  )
})
