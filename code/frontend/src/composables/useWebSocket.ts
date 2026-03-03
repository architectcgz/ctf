import { onUnmounted, ref } from 'vue'

import { getWsTicket } from '@/api/auth'
import { WS_MAX_RECONNECT_ATTEMPTS, WS_HEARTBEAT_INTERVAL_MS } from '@/utils/constants'

export type WebSocketStatus = 'idle' | 'connecting' | 'open' | 'closed' | 'error'

export interface WebSocketHandlers {
  [type: string]: (payload: unknown) => void
}

export function useWebSocket(endpoint: string, handlers: WebSocketHandlers) {
  const status = ref<WebSocketStatus>('idle')

  let socket: WebSocket | null = null
  let heartbeatTimer: number | undefined
  let reconnectTimer: number | undefined
  let reconnectAttempt = 0

  const wsBase = import.meta.env.VITE_WS_BASE_URL || '/ws'

  function resolveWsUrl(pathWithQuery: string): string {
    if (/^wss?:\/\//i.test(pathWithQuery)) return pathWithQuery
    const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:'
    const path = pathWithQuery.startsWith('/') ? pathWithQuery : `/${pathWithQuery}`
    return `${protocol}//${window.location.host}${path}`
  }

  function clearTimers(): void {
    if (heartbeatTimer) window.clearInterval(heartbeatTimer)
    if (reconnectTimer) window.clearTimeout(reconnectTimer)
    heartbeatTimer = undefined
    reconnectTimer = undefined
  }

  async function connect(): Promise<void> {
    clearTimers()
    status.value = 'connecting'

    const { ticket } = await getWsTicket()
    const path = `${wsBase.replace(/\/$/, '')}/${endpoint.replace(/^\//, '')}`
    socket = new WebSocket(resolveWsUrl(path))

    socket.addEventListener('open', () => {
      status.value = 'open'
      reconnectAttempt = 0
      // 连接建立后立即发送认证消息
      socket?.send(JSON.stringify({ type: 'auth', payload: { ticket }, timestamp: new Date().toISOString() }))
      heartbeatTimer = window.setInterval(() => {
        try {
          socket?.send(JSON.stringify({ type: 'ping', payload: {}, timestamp: new Date().toISOString() }))
        } catch {
          // ignore
        }
      }, WS_HEARTBEAT_INTERVAL_MS)
    })

    socket.addEventListener('message', (evt) => {
      try {
        const msg = JSON.parse(String(evt.data)) as { type?: string; payload?: unknown; timestamp?: string }
        const type = msg.type
        if (!type) return
        const handler = handlers[type]
        if (handler) handler(msg.payload)
      } catch (err) {
        console.error('WS message parse failed:', err)
      }
    })

    socket.addEventListener('close', () => {
      status.value = 'closed'
      clearTimers()
      scheduleReconnect()
    })

    socket.addEventListener('error', () => {
      status.value = 'error'
    })
  }

  function scheduleReconnect(): void {
    if (reconnectAttempt >= WS_MAX_RECONNECT_ATTEMPTS) return
    reconnectAttempt += 1
    const delayMs = Math.min(30_000, 1000 * 2 ** (reconnectAttempt - 1))
    reconnectTimer = window.setTimeout(() => {
      connect().catch((err) => {
        console.error('WS reconnect failed:', err)
        scheduleReconnect()
      })
    }, delayMs)
  }

  function disconnect(): void {
    clearTimers()
    socket?.close()
    socket = null
    status.value = 'closed'
  }

  function send(data: unknown): void {
    if (status.value !== 'open') return
    socket?.send(JSON.stringify(data))
  }

  onUnmounted(() => {
    disconnect()
  })

  return { status, connect, disconnect, send }
}
