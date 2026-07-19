import { useState, useEffect, useRef } from 'react'
import { Box, Typography, CircularProgress } from '@mui/material'
import { getLogsByJob } from '../lib/api'
import type { JobLog } from '../lib/types'

interface Props {
  jobId: string
  /** New log chunks pushed in from the parent's SSE handler */
  liveChunks: JobLog[]
}

export default function JobLogsPanel({ jobId, liveChunks }: Props) {
  const [lines, setLines] = useState<string[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')
  const bottomRef = useRef<HTMLDivElement>(null)

  // Initial load of existing log chunks
  useEffect(() => {
    setLoading(true)
    setError('')
    getLogsByJob(jobId)
      .then((chunks) => {
        const all = chunks.flatMap((c) => c.log_data.split('\n').filter(Boolean))
        setLines(all)
      })
      .catch((e) => setError(e instanceof Error ? e.message : 'Failed to load logs'))
      .finally(() => setLoading(false))
  }, [jobId])

  // Append new live chunks as they arrive via SSE — track processed count to avoid re-appending
  const processedCountRef = useRef(0)
  useEffect(() => {
    const unprocessed = liveChunks.slice(processedCountRef.current)
    if (unprocessed.length === 0) return
    const newLines = unprocessed.flatMap((c) => c.log_data.split('\n').filter(Boolean))
    setLines((prev) => [...prev, ...newLines])
    processedCountRef.current = liveChunks.length
  }, [liveChunks])

  // Auto-scroll to bottom on new lines
  useEffect(() => {
    bottomRef.current?.scrollIntoView({ behavior: 'smooth' })
  }, [lines])

  return (
    <Box
      sx={{
        bgcolor: '#080a0e',
        border: '1px solid #1e2028',
        borderTop: 'none',
        fontFamily: 'monospace',
        fontSize: '0.72rem',
        lineHeight: 1.6,
        maxHeight: 300,
        overflowY: 'auto',
        px: 2,
        py: 1.5,
      }}
    >
      {loading && (
        <Box sx={{ display: 'flex', alignItems: 'center', gap: 1, color: '#334155' }}>
          <CircularProgress size={10} thickness={4} sx={{ color: '#334155' }} />
          <span>Loading logs…</span>
        </Box>
      )}

      {error && (
        <Typography sx={{ color: '#f87171', fontSize: '0.72rem' }}>{error}</Typography>
      )}

      {!loading && !error && lines.length === 0 && (
        <Typography sx={{ color: '#334155', fontSize: '0.72rem' }}>No logs yet.</Typography>
      )}

      {lines.map((line, i) => {
        // Lines are formatted as "[2026-07-11T12:34:01Z] message"
        const match = line.match(/^\[([^\]]+)\]\s(.*)$/)
        const ts = match?.[1] ?? ''
        const msg = match?.[2] ?? line
        return (
          <Box key={i} sx={{ display: 'flex', gap: 2, whiteSpace: 'pre-wrap', wordBreak: 'break-all' }}>
            <Box component="span" sx={{ color: '#334155', flexShrink: 0 }}>
              {ts ? new Date(ts).toLocaleTimeString() : ''}
            </Box>
            <Box component="span" sx={{ color: '#94a3b8' }}>{msg}</Box>
          </Box>
        )
      })}

      <div ref={bottomRef} />
    </Box>
  )
}
