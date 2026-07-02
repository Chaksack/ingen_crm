import { defineStore } from 'pinia'

export interface ChatMessage {
  id: string
  sender_user_id: string
  recipient_user_id: string
  body: string
  created_at: string
}

export const useChatStore = defineStore('chat', {
  state: () => ({
    socket: null as WebSocket | null,
    connected: false,
    messagesByPeer: {} as Record<string, ChatMessage[]>,
  }),
  actions: {
    connect() {
      if (this.socket || !import.meta.client) return
      const config = useRuntimeConfig()
      const auth = useAuthStore()
      if (!auth.token) return

      const socket = new WebSocket(`${config.public.wsBase}/ws?token=${auth.token}`)
      socket.addEventListener('open', () => {
        this.connected = true
      })
      socket.addEventListener('close', () => {
        this.connected = false
        this.socket = null
      })
      socket.addEventListener('message', (event) => {
        const msg = JSON.parse(event.data) as ChatMessage & { type: string }
        if (msg.type !== 'message') return
        const auth = useAuthStore()
        const peer = msg.sender_user_id === auth.user?.id ? msg.recipient_user_id : msg.sender_user_id
        if (!this.messagesByPeer[peer]) this.messagesByPeer[peer] = []
        this.messagesByPeer[peer].push(msg)
      })
      this.socket = socket
    },
    disconnect() {
      this.socket?.close()
      this.socket = null
      this.connected = false
    },
    send(to: string, body: string) {
      if (!this.socket || this.socket.readyState !== WebSocket.OPEN) return
      this.socket.send(JSON.stringify({ to, body }))
    },
    setHistory(peer: string, messages: ChatMessage[]) {
      this.messagesByPeer[peer] = messages
    },
  },
})
