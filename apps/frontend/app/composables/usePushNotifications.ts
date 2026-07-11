// Web Push (PRD §5.11/§7.5): subscribes the current browser/device to
// receive pushes even when the PWA is closed. The service worker's `push`
// and `notificationclick` handlers live in app/service-worker/sw.ts.

function urlBase64ToUint8Array(base64: string) {
  const padding = '='.repeat((4 - (base64.length % 4)) % 4)
  const base64Safe = (base64 + padding).replace(/-/g, '+').replace(/_/g, '/')
  const raw = atob(base64Safe)
  return Uint8Array.from([...raw].map((c) => c.charCodeAt(0)))
}

export function usePushNotifications() {
  const api = useApi()
  const supported = import.meta.client && 'serviceWorker' in navigator && 'PushManager' in window
  const permission = ref<NotificationPermission>(supported ? Notification.permission : 'denied')
  const subscribed = ref(false)
  const error = ref('')

  async function refreshStatus() {
    if (!supported) return
    const reg = await navigator.serviceWorker.ready
    const sub = await reg.pushManager.getSubscription()
    subscribed.value = !!sub
    permission.value = Notification.permission
  }

  async function subscribe() {
    error.value = ''
    if (!supported) {
      error.value = 'Push notifications are not supported in this browser.'
      return
    }
    try {
      const perm = await Notification.requestPermission()
      permission.value = perm
      if (perm !== 'granted') {
        error.value = 'Notification permission was not granted.'
        return
      }
      const { public_key: publicKey } = await api<{ public_key: string }>('/push/vapid-public-key')
      if (!publicKey) {
        error.value = 'Push is not configured on the server.'
        return
      }
      const reg = await navigator.serviceWorker.ready
      const sub = await reg.pushManager.subscribe({
        userVisibleOnly: true,
        applicationServerKey: urlBase64ToUint8Array(publicKey),
      })
      const json = sub.toJSON()
      await api('/push/subscribe', {
        method: 'POST',
        body: { endpoint: json.endpoint, keys: json.keys },
      })
      subscribed.value = true
    } catch {
      error.value = 'Could not enable push notifications.'
    }
  }

  async function unsubscribe() {
    if (!supported) return
    try {
      const reg = await navigator.serviceWorker.ready
      const sub = await reg.pushManager.getSubscription()
      if (sub) {
        await api('/push/subscribe', { method: 'DELETE', body: { endpoint: sub.endpoint } })
        await sub.unsubscribe()
      }
      subscribed.value = false
    } catch {
      error.value = 'Could not disable push notifications.'
    }
  }

  if (supported) {
    onMounted(refreshStatus)
  }

  return { supported, permission, subscribed, error, subscribe, unsubscribe }
}
