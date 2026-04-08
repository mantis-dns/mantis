import { useEffect, useRef, useState, useCallback } from 'react'

export function useWebSocket<T>(url: string, enabled = true) {
  const [messages, setMessages] = useState<T[]>([])
  const [connected, setConnected] = useState(false)
  const wsRef = useRef<WebSocket | null>(null)
  const reconnectTimer = useRef<ReturnType<typeof setTimeout>>(undefined)

  const connect = useCallback(() => {
    if (!enabled) return

    const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:'
    const ws = new WebSocket(`${protocol}//${window.location.host}${url}`)
    wsRef.current = ws

    ws.onopen = () => setConnected(true)
    ws.onclose = () => {
      setConnected(false)
      reconnectTimer.current = setTimeout(connect, 3000)
    }
    ws.onmessage = (e) => {
      const data = JSON.parse(e.data) as T
      setMessages((prev) => [data, ...prev.slice(0, 999)])
    }
  }, [url, enabled])

  useEffect(() => {
    connect()
    return () => {
      clearTimeout(reconnectTimer.current)
      wsRef.current?.close()
    }
  }, [connect])

  const clear = useCallback(() => setMessages([]), [])

  return { messages, connected, clear }
}
