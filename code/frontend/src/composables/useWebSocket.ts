import { onUnmounted, ref } from 'vue'

import { getWsTicket } from '@/api/auth'
import { useAuthStore } from '@/stores/auth'
import {
  WS_HEARTBEAT_INTERVAL_MS,
  WS_MAX_RECONNECT_ATTEMPTS,
  WS_MAX_RECONNECT_DELAY_MS,
  WS_PONG_TIMEOUT_MS,
  WS_RECONNECT_BASE_DELAY_MS,
} from '@/utils/constants'
import { redirectToErrorStatusPage } from '@/utils/errorStatusPage'

export type WebSocketStatus = 'idle' | 'connecting' | 'open' | 'closed' | 'error'

export interface WebSocketHandlers {
  [type: string]: (payload: unknown) => void
}

const AUTH_CLOSE_CODES = new Set([4001, 4401])

export function useWebSocket(endpoint: string, handlers: WebSocketHandlers) {
  const status = ref<WebSocketStatus>('idle')

  let socket: WebSocket | null = null
  let heartbeatTimer: number | undefined
  let reconnectTimer: number | undefined
  let pongTimeoutTimer: number | undefined
  let reconnectAttempt = 0
  let manualClose = false

  const wsBase = import.meta.env.VITE_WS_BASE_URL || '/ws'

  function clearTimers(): void {
    if (heartbeatTimer) window.clearInterval(heartbeatTimer)
    if (reconnectTimer) window.clearTimeout(reconnectTimer)
    if (pongTimeoutTimer) window.clearTimeout(pongTimeoutTimer)
    heartbeatTimer = undefined
    reconnectTimer = undefined
    pongTimeoutTimer = undefined
  }

  function resolveWsUrl(pathWithQuery: string): string {
    if (/^wss?:\/\//i.test(pathWithQuery)) return pathWithQuery
    const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:'
    const normalizedPath = pathWithQuery.startsWith('/') ? pathWithQuery : `/${pathWithQuery}`
    return `${protocol}//${window.location.host}${normalizedPath}`
  }

  function buildWsPath(ticket: string): string {
    const base = wsBase.replace(/\/$/, '')
    const target = endpoint.replace(/^\//, '')
    const separator = target.includes('?') ? '&' : '?'
    return `${base}/${target}${separator}ticket=${encodeURIComponent(ticket)}`
  }

  function clearPongTimeout(): void {
    if (pongTimeoutTimer) {
      window.clearTimeout(pongTimeoutTimer)
      pongTimeoutTimer = undefined
    }
  }

  function startPongTimeout(): void {
    clearPongTimeout()
    pongTimeoutTimer = window.setTimeout(() => {
      if (socket && socket.readyState === WebSocket.OPEN) {
        socket.close(4000, 'pong-timeout')
      }
    }, WS_PONG_TIMEOUT_MS)
  }

  function startHeartbeat(): void {
    heartbeatTimer = window.setInterval(() => {
      if (!socket || socket.readyState !== WebSocket.OPEN) return
      socket.send(
        JSON.stringify({ type: 'ping', payload: {}, timestamp: new Date().toISOString() })
      )
      startPongTimeout()
    }, WS_HEARTBEAT_INTERVAL_MS)
  }

  function handleAuthClosed(): void {
    const authStore = useAuthStore()
    authStore.logout()
    redirectToErrorStatusPage(401)
  }

  function scheduleReconnect(): void {
    if (manualClose) return
    if (reconnectAttempt >= WS_MAX_RECONNECT_ATTEMPTS) return

    reconnectAttempt += 1
    const delayMs = Math.min(
      WS_MAX_RECONNECT_DELAY_MS,
      WS_RECONNECT_BASE_DELAY_MS * 2 ** (reconnectAttempt - 1)
    )
    reconnectTimer = window.setTimeout(() => {
      connect().catch((err) => {
        console.error('WS reconnect failed:', err)
        scheduleReconnect()
      })
    }, delayMs)
  }

  async function connect(): Promise<void> {
    if (status.value === 'connecting' || status.value === 'open') return
    manualClose = false
    clearTimers()
    status.value = 'connecting'

    try {
      const { ticket } = await getWsTicket()
      const path = buildWsPath(ticket)
      socket = new WebSocket(resolveWsUrl(path))

      socket.addEventListener('open', () => {
        status.value = 'open'
        reconnectAttempt = 0
        startHeartbeat()
      })

      socket.addEventListener('message', (evt) => {
        try {
          const msg = JSON.parse(String(evt.data)) as { type?: string; payload?: unknown }
          const type = msg.type
          if (!type) return
          if (type === 'pong') {
            clearPongTimeout()
            return
          }
          handlers[type]?.(msg.payload)
        } catch (err) {
          console.error('WS message parse failed:', err)
        }
      })

      socket.addEventListener('close', (event) => {
        status.value = 'closed'
        clearTimers()
        socket = null

        if (manualClose) return
        if (AUTH_CLOSE_CODES.has(event.code)) {
          handleAuthClosed()
          return
        }
        scheduleReconnect()
      })

      socket.addEventListener('error', () => {
        status.value = 'error'
      })
    } catch (err) {
      status.value = 'error'
      scheduleReconnect()
      throw err
    }
  }

  function disconnect(): void {
    manualClose = true
    clearTimers()
    socket?.close(1000, 'manual-close')
    socket = null
    status.value = 'closed'
  }

  function send(data: unknown): void {
    if (!socket || socket.readyState !== WebSocket.OPEN) return
    socket.send(JSON.stringify(data))
  }

  onUnmounted(() => {
    disconnect()
  })

  return { status, connect, disconnect, send }
}
