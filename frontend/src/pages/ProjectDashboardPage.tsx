import { useState, useEffect, useCallback } from 'react'
import { useParams, useNavigate } from 'react-router-dom'
import {
  Box,
  Typography,
  Button,
  Alert,
  Divider,
  Table,
  TableBody,
  TableCell,
  TableContainer,
  TableHead,
  TableRow,
  Tooltip,
  Collapse,
  IconButton,
} from '@mui/material'
import { KeyboardArrowDownRounded, KeyboardArrowUpRounded } from '@mui/icons-material'
import { getProjectDetails, getJobsByProject } from '../lib/api'
import { useSSE } from '../hooks/useSSE'
import type { ProjectDetails, Job, JobLog, SSEDatabaseEvent } from '../lib/types'
import JobStatusBadge from '../components/JobStatusBadge'
import JobLogsPanel from '../components/JobLogsPanel'
import Spinner from '../components/Spinner'

function formatDuration(start: string, end: string | null): string {
  if (!end) return '—'
  const ms = new Date(end).getTime() - new Date(start).getTime()
  if (ms < 0) return '—'
  const s = Math.floor(ms / 1000)
  if (s < 60) return `${s}s`
  return `${Math.floor(s / 60)}m ${s % 60}s`
}

function formatRelative(iso: string): string {
  const diff = Date.now() - new Date(iso).getTime()
  const s = Math.floor(diff / 1000)
  if (s < 60) return `${s}s ago`
  const m = Math.floor(s / 60)
  if (m < 60) return `${m}m ago`
  const h = Math.floor(m / 60)
  if (h < 24) return `${h}h ago`
  return `${Math.floor(h / 24)}d ago`
}

function Stat({ label, value, color = '#e2e8f0' }: { label: string; value: number; color?: string }) {
  return (
    <Box>
      <Typography sx={{ fontSize: '1.4rem', fontWeight: 500, color, lineHeight: 1 }}>
        {value}
      </Typography>
      <Typography variant="caption" sx={{ color: '#334155', mt: 0.25, display: 'block' }}>
        {label}
      </Typography>
    </Box>
  )
}

// ─── Job row with expandable log panel ───────────────────────────────────────

interface JobRowProps {
  job: Job
  liveLogChunks: JobLog[]
}

function JobRow({ job, liveLogChunks }: JobRowProps) {
  const [open, setOpen] = useState(false)

  return (
    <>
      <TableRow>
        <TableCell sx={{ py: 1, width: 32, pr: 0 }}>
          <IconButton
            size="small"
            onClick={() => setOpen((s) => !s)}
            sx={{ color: '#334155', '&:hover': { color: '#94a3b8' } }}
          >
            {open
              ? <KeyboardArrowUpRounded fontSize="small" />
              : <KeyboardArrowDownRounded fontSize="small" />
            }
          </IconButton>
        </TableCell>

        <TableCell sx={{ py: 1.5 }}>
          <JobStatusBadge status={job.status} />
        </TableCell>

        <TableCell>
          <Tooltip title={job.commit_sha ?? ''} placement="top" arrow>
            <Typography
              component="code"
              sx={{ fontFamily: 'monospace', fontSize: '0.75rem', color: '#64748b', cursor: 'default' }}
            >
              {(job.commit_sha ?? '').slice(0, 7) || '—'}
            </Typography>
          </Tooltip>
        </TableCell>

        <TableCell>
          <Typography variant="body2" sx={{ color: '#475569' }} noWrap>
            {job.repo_url ?? '—'}
          </Typography>
        </TableCell>

        <TableCell>
          <Typography variant="body2" sx={{ color: '#475569' }}>
            {job.created_at ? formatRelative(job.created_at) : '—'}
          </Typography>
        </TableCell>

        <TableCell align="right">
          <Typography sx={{ fontFamily: 'monospace', fontSize: '0.75rem', color: '#334155' }}>
            {formatDuration(job.created_at, job.finished_at)}
          </Typography>
        </TableCell>
      </TableRow>

      {/* Log panel row */}
      <TableRow sx={{ '&:hover td': { bgcolor: 'transparent !important' } }}>
        <TableCell colSpan={6} sx={{ p: 0, border: open ? undefined : 'none' }}>
          <Collapse in={open}>
            <JobLogsPanel jobId={job.job_id} liveChunks={liveLogChunks} />
          </Collapse>
        </TableCell>
      </TableRow>
    </>
  )
}

// ─── Page ─────────────────────────────────────────────────────────────────────

export default function ProjectDashboardPage() {
  const { projectId } = useParams<{ projectId: string }>()
  const navigate = useNavigate()

  const [project, setProject] = useState<ProjectDetails | null>(null)
  const [jobs, setJobs] = useState<Job[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')
  const [liveIndicator, setLiveIndicator] = useState(false)

  // Map of job_id → live log chunks received via SSE
  const [liveLogChunks, setLiveLogChunks] = useState<Record<string, JobLog[]>>({})

  const flashLive = useCallback(() => {
    setLiveIndicator(true)
    setTimeout(() => setLiveIndicator(false), 1500)
  }, [])

  useEffect(() => {
    if (!projectId) return
    setLoading(true)
    Promise.all([getProjectDetails(projectId), getJobsByProject(projectId)])
      .then(([proj, jobList]) => { setProject(proj); setJobs(jobList) })
      .catch((e) => setError(e instanceof Error ? e.message : 'Failed to load project'))
      .finally(() => setLoading(false))
  }, [projectId])

  const handleSSEEvent = useCallback(
    (ev: SSEDatabaseEvent) => {
      flashLive()

      if (ev.entity_name === 'job') {
        const incoming = ev.entity_data as Partial<Job> & { job_id?: string }
        const jobId = incoming.job_id
        if (!jobId) return

        if (ev.method === 'create') {
          setJobs((prev) => {
            if (prev.some((j) => j.job_id === jobId)) return prev
            return [{
              job_id: jobId,
              project_id: incoming.project_id ?? projectId ?? null,
              repo_url: incoming.repo_url ?? null,
              commit_sha: incoming.commit_sha ?? '',
              created_at: incoming.created_at ?? new Date().toISOString(),
              finished_at: incoming.finished_at ?? null,
              status: incoming.status ?? null,
            }, ...prev]
          })
        } else if (ev.method === 'update') {
          setJobs((prev) => prev.map((j) => j.job_id === jobId ? { ...j, ...incoming } : j))
        } else if (ev.method === 'delete') {
          setJobs((prev) => prev.filter((j) => j.job_id !== jobId))
        }
      }

      if (ev.entity_name === 'job_log' && ev.method === 'create') {
        const chunk = ev.entity_data as Partial<JobLog> & { job_id?: string }
        if (!chunk.job_id || !chunk.log_data) return
        const newChunk: JobLog = {
          log_id: (chunk as JobLog).log_id ?? '',
          job_id: chunk.job_id,
          log_data: chunk.log_data,
          created_at: (chunk as JobLog).created_at ?? new Date().toISOString(),
        }
        setLiveLogChunks((prev) => ({
          ...prev,
          [chunk.job_id!]: [...(prev[chunk.job_id!] ?? []), newChunk],
        }))
      }
    },
    [projectId, flashLive],
  )

  useSSE(projectId ?? null, handleSSEEvent)

  if (loading) return <Spinner centered />

  if (error) {
    return (
      <Box>
        <Alert severity="error" sx={{ mb: 2 }}>{error}</Alert>
        <Button variant="text" size="small" onClick={() => navigate('/projects')}>← Projects</Button>
      </Box>
    )
  }

  const passing = jobs.filter((j) => j.status === 'Passed').length
  const failing = jobs.filter((j) => j.status === 'Failed').length
  const running = jobs.filter((j) => j.status === 'Running').length

  return (
    <Box>
      {/* Breadcrumb */}
      <Box sx={{ display: 'flex', alignItems: 'center', gap: 1, mb: 3 }}>
        <Typography
          variant="caption"
          sx={{ color: '#334155', cursor: 'pointer', '&:hover': { color: '#64748b' } }}
          onClick={() => navigate('/projects')}
        >
          Projects
        </Typography>
        <Typography variant="caption" sx={{ color: '#1e2028' }}>/</Typography>
        <Typography variant="caption" sx={{ color: '#64748b' }}>{project?.name}</Typography>
      </Box>

      {/* Project meta */}
      <Box sx={{ mb: 3 }}>
        <Box sx={{ display: 'flex', alignItems: 'flex-start', justifyContent: 'space-between', gap: 2 }}>
          <Box>
            <Typography variant="h4" sx={{ color: '#e2e8f0' }}>{project?.name}</Typography>
            <Typography
              component="a"
              href={project?.repo_url}
              target="_blank"
              rel="noopener noreferrer"
              variant="body2"
              sx={{ color: '#475569', textDecoration: 'none', '&:hover': { color: '#94a3b8' }, mt: 0.5, display: 'block' }}
            >
              {project?.repo_url}
            </Typography>
          </Box>
          <Box sx={{ display: 'flex', alignItems: 'center', gap: 1, mt: 0.5 }}>
            <Box sx={{ width: 6, height: 6, borderRadius: '50%', bgcolor: liveIndicator ? '#4ade80' : '#1e2028', transition: 'background-color 0.3s' }} />
            <Typography variant="caption" sx={{ color: '#334155' }}>live</Typography>
          </Box>
        </Box>
        <Typography variant="caption" sx={{ color: '#1e2028', mt: 1, display: 'block' }}>
          {project?.owner_username} · {project?.owner_email}
        </Typography>
      </Box>

      <Divider />

      {/* Stats */}
      <Box sx={{ display: 'flex', gap: 4, py: 2.5, borderBottom: '1px solid #1e2028' }}>
        <Stat label="total" value={jobs.length} />
        <Stat label="passed" value={passing} color="#4ade80" />
        <Stat label="failed" value={failing} color="#f87171" />
        <Stat label="running" value={running} color="#60a5fa" />
      </Box>

      {/* Runs */}
      <Box sx={{ mt: 3, mb: 1, display: 'flex', alignItems: 'center', justifyContent: 'space-between' }}>
        <Typography variant="subtitle2" sx={{ color: '#64748b' }}>Runs</Typography>
        <Typography variant="caption" sx={{ color: '#1e2028' }}>{jobs.length} total</Typography>
      </Box>

      {jobs.length === 0 ? (
        <Box sx={{ py: 8, textAlign: 'center', borderTop: '1px solid #1e2028' }}>
          <Typography variant="body2" sx={{ color: '#334155' }}>
            No runs yet. Push a commit to trigger the first pipeline.
          </Typography>
        </Box>
      ) : (
        <TableContainer sx={{ border: '1px solid #1e2028', borderRadius: '2px' }}>
          <Table size="small">
            <TableHead>
              <TableRow>
                <TableCell width={32} sx={{ pr: 0 }} />
                <TableCell width={90}>Status</TableCell>
                <TableCell width={90}>Commit</TableCell>
                <TableCell>Repository</TableCell>
                <TableCell width={110}>Triggered</TableCell>
                <TableCell width={80} align="right">Duration</TableCell>
              </TableRow>
            </TableHead>
            <TableBody>
              {jobs.map((job) => (
                <JobRow
                  key={job.job_id}
                  job={job}
                  liveLogChunks={liveLogChunks[job.job_id] ?? []}
                />
              ))}
            </TableBody>
          </Table>
        </TableContainer>
      )}
    </Box>
  )
}
