import { useEffect, useRef, useCallback } from 'react'
import { sseUrl } from '../lib/api'
import type { SSEDatabaseEvent } from '../lib/types'

type SSEHandler = (event: SSEDatabaseEvent) => void

/**
 * Connects to the DB-service SSE stream for a given projectId.
 * Automatically reconnects on error with exponential back-off.
 * Cleans up on unmount or when projectId changes.
 */
export function useSSE(projectId: string | null, onEvent: SSEHandler) {
  const handlerRef = useRef<SSEHandler>(onEvent)
  handlerRef.current = onEvent

  const retryDelay = useRef(1000)
  const esRef = useRef<EventSource | null>(null)
  const timerRef = useRef<ReturnType<typeof setTimeout> | null>(null)

  const connect = useCallback(() => {
    if (!projectId) return

    const es = new EventSource(sseUrl(projectId))
    esRef.current = es

    es.addEventListener('database-event', (e: MessageEvent) => {
      retryDelay.current = 1000
      try {
        const data = JSON.parse(e.data) as SSEDatabaseEvent
        handlerRef.current(data)
      } catch {
        console.error('[SSE] failed to parse event', e.data)
      }
    })

    es.onerror = () => {
      es.close()
      esRef.current = null
      timerRef.current = setTimeout(() => {
        retryDelay.current = Math.min(retryDelay.current * 2, 30_000)
        connect()
      }, retryDelay.current)
    }
  }, [projectId])

  useEffect(() => {
    if (!projectId) return

    retryDelay.current = 1000
    connect()

    return () => {
      esRef.current?.close()
      esRef.current = null
      if (timerRef.current) clearTimeout(timerRef.current)
    }
  }, [projectId, connect])
}
