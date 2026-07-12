import { useState, useEffect, type FormEvent } from 'react'
import { useNavigate } from 'react-router-dom'
import { v4 as uuidv4 } from 'uuid'
import {
  Box,
  Button,
  Collapse,
  TextField,
  Typography,
  Alert,
  Skeleton,
  Divider,
  Grid,
} from '@mui/material'
import { useAuth } from '../context/AuthContext'
import { getProjectsByUser, createProject } from '../lib/api'
import type { Project } from '../lib/types'

export default function ProjectsPage() {
  const { user } = useAuth()
  const navigate = useNavigate()

  const [projects, setProjects] = useState<Project[]>([])
  const [loading, setLoading] = useState(true)
  const [fetchError, setFetchError] = useState('')

  const [showForm, setShowForm] = useState(false)
  const [name, setName] = useState('')
  const [repoUrl, setRepoUrl] = useState('')
  const [creating, setCreating] = useState(false)
  const [createError, setCreateError] = useState('')

  useEffect(() => {
    if (!user) return
    setLoading(true)
    getProjectsByUser(user.user_id)
      .then(setProjects)
      .catch((e) => setFetchError(e instanceof Error ? e.message : 'Failed to load projects'))
      .finally(() => setLoading(false))
  }, [user])

  async function handleCreate(e: FormEvent) {
    e.preventDefault()
    if (!user) return
    setCreateError('')
    setCreating(true)
    try {
      await createProject(uuidv4(), user.user_id, name.trim(), repoUrl.trim())
      const updated = await getProjectsByUser(user.user_id)
      setProjects(updated)
      setShowForm(false)
      setName('')
      setRepoUrl('')
    } catch (err) {
      setCreateError(err instanceof Error ? err.message : 'Failed to create project')
    } finally {
      setCreating(false)
    }
  }

  return (
    <Box>
      {/* Page header */}
      <Box sx={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', mb: 3 }}>
        <Box>
          <Typography variant="h4" sx={{ color: '#e2e8f0' }}>
            Projects
          </Typography>
          <Typography variant="body2" sx={{ color: '#475569', mt: 0.5 }}>
            {loading ? '' : `${projects.length} repositories`}
          </Typography>
        </Box>
        <Button
          variant="contained"
          size="small"
          onClick={() => { setShowForm((s) => !s); setCreateError('') }}
        >
          New project
        </Button>
      </Box>

      <Divider />

      {/* New project form */}
      <Collapse in={showForm} unmountOnExit>
        <Box sx={{ py: 3, borderBottom: '1px solid #1e2028' }}>
          <Typography variant="subtitle2" sx={{ color: '#94a3b8', mb: 2 }}>
            New project
          </Typography>

          {createError && <Alert severity="error" sx={{ mb: 2 }}>{createError}</Alert>}

          <Box component="form" onSubmit={handleCreate}>
            <Grid container spacing={2} sx={{ mb: 2 }}>
              <Grid size={{ xs: 12, sm: 5 }}>
                <TextField
                  label="Project name"
                  required
                  fullWidth
                  value={name}
                  onChange={(e) => setName(e.target.value)}
                  placeholder="my-service"
                />
              </Grid>
              <Grid size={{ xs: 12, sm: 7 }}>
                <TextField
                  label="Repository URL"
                  type="url"
                  required
                  fullWidth
                  value={repoUrl}
                  onChange={(e) => setRepoUrl(e.target.value)}
                  placeholder="https://github.com/org/repo"
                />
              </Grid>
            </Grid>
            <Box sx={{ display: 'flex', gap: 1 }}>
              <Button type="submit" variant="contained" size="small" disabled={creating}>
                {creating ? 'Creating…' : 'Create'}
              </Button>
              <Button variant="text" size="small" onClick={() => setShowForm(false)}>
                Cancel
              </Button>
            </Box>
          </Box>
        </Box>
      </Collapse>

      {/* Error */}
      {fetchError && <Alert severity="error" sx={{ mt: 3 }}>{fetchError}</Alert>}

      {/* Loading */}
      {loading && (
        <Box sx={{ mt: 1 }}>
          {[1, 2, 3, 4].map((i) => (
            <Box key={i} sx={{ borderBottom: '1px solid #1e2028' }}>
              <Skeleton height={52} />
            </Box>
          ))}
        </Box>
      )}

      {/* Empty */}
      {!loading && !fetchError && projects.length === 0 && (
        <Box sx={{ py: 8, textAlign: 'center' }}>
          <Typography variant="body2" sx={{ color: '#334155' }}>
            No projects. Create one to get started.
          </Typography>
        </Box>
      )}

      {/* Project rows */}
      {!loading && projects.length > 0 && (
        <Box>
          {projects.map((p) => (
            <Box
              key={p.project_id}
              onClick={() => navigate(`/projects/${p.project_id}`)}
              sx={{
                display: 'flex',
                alignItems: 'center',
                justifyContent: 'space-between',
                px: 0,
                py: 1.75,
                borderBottom: '1px solid #1e2028',
                cursor: 'pointer',
                '&:hover': {
                  '& .proj-name': { color: '#f8fafc' },
                },
              }}
            >
              <Box sx={{ minWidth: 0 }}>
                <Typography
                  className="proj-name"
                  variant="body2"
                  sx={{ color: '#cbd5e1', fontWeight: 500, transition: 'color 0.15s' }}
                  noWrap
                >
                  {p.name}
                </Typography>
                <Typography variant="caption" sx={{ color: '#334155' }} noWrap>
                  {p.repo_url}
                </Typography>
              </Box>
              <Typography
                variant="caption"
                sx={{
                  color: '#1e2028',
                  fontFamily: 'monospace',
                  ml: 2,
                  flexShrink: 0,
                  display: { xs: 'none', sm: 'block' },
                }}
              >
                {p.project_id.slice(0, 8)}
              </Typography>
            </Box>
          ))}
        </Box>
      )}
    </Box>
  )
}
