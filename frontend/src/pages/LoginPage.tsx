import { useState, type FormEvent } from 'react'
import { useNavigate, Link } from 'react-router-dom'
import {
  Box,
  Button,
  TextField,
  Typography,
  Alert,
  InputAdornment,
  IconButton,
  Divider,
} from '@mui/material'
import { Visibility, VisibilityOff } from '@mui/icons-material'
import { useAuth } from '../context/AuthContext'
import { signIn } from '../lib/api'

export default function LoginPage() {
  const { login } = useAuth()
  const navigate = useNavigate()

  const [email, setEmail] = useState('')
  const [password, setPassword] = useState('')
  const [showPassword, setShowPassword] = useState(false)
  const [error, setError] = useState('')
  const [loading, setLoading] = useState(false)

  async function handleSubmit(e: FormEvent) {
    e.preventDefault()
    setError('')
    setLoading(true)
    try {
      const user = await signIn(email.trim(), password)
      login(user)
      navigate('/projects')
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Sign in failed')
    } finally {
      setLoading(false)
    }
  }

  return (
    <Box
      sx={{
        minHeight: '100vh',
        display: 'flex',
        alignItems: 'center',
        justifyContent: 'center',
        bgcolor: 'background.default',
        px: 2,
      }}
    >
      <Box sx={{ width: '100%', maxWidth: 360 }}>
        {/* Wordmark */}
        <Box sx={{ display: 'flex', alignItems: 'center', gap: 1.5, mb: 5 }}>
          <Box sx={{ width: 20, height: 20, bgcolor: '#e2e8f0', display: 'flex', alignItems: 'center', justifyContent: 'center' }}>
            <Box component="span" sx={{ display: 'block', width: 8, height: 8, bgcolor: '#0c0e12' }} />
          </Box>
          <Typography variant="body2" sx={{ color: '#e2e8f0', fontWeight: 500, letterSpacing: '0.02em' }}>
            braceletci
          </Typography>
        </Box>

        <Typography variant="h5" sx={{ color: '#e2e8f0', mb: 0.5 }}>
          Sign in
        </Typography>
        <Typography variant="body2" sx={{ color: '#475569', mb: 4 }}>
          Enter your credentials to continue
        </Typography>

        {error && (
          <Alert severity="error" sx={{ mb: 3 }}>
            {error}
          </Alert>
        )}

        <Box component="form" onSubmit={handleSubmit} sx={{ display: 'flex', flexDirection: 'column', gap: 2 }}>
          <TextField
            label="Email"
            type="email"
            required
            fullWidth
            value={email}
            onChange={(e) => setEmail(e.target.value)}
            autoComplete="email"
          />

          <TextField
            label="Password"
            type={showPassword ? 'text' : 'password'}
            required
            fullWidth
            value={password}
            onChange={(e) => setPassword(e.target.value)}
            autoComplete="current-password"
            slotProps={{
              input: {
                endAdornment: (
                  <InputAdornment position="end">
                    <IconButton size="small" onClick={() => setShowPassword((s) => !s)} edge="end">
                      {showPassword ? <VisibilityOff fontSize="small" /> : <Visibility fontSize="small" />}
                    </IconButton>
                  </InputAdornment>
                ),
              },
            }}
          />

          <Button
            type="submit"
            variant="contained"
            fullWidth
            disabled={loading}
            sx={{ mt: 1, py: '7px' }}
          >
            {loading ? 'Signing in…' : 'Sign in'}
          </Button>
        </Box>

        <Divider sx={{ my: 3, borderColor: '#1e2028' }} />

        <Typography variant="body2" sx={{ color: '#475569' }}>
          No account?{' '}
          <Link to="/register" style={{ color: '#94a3b8', textDecoration: 'none' }}>
            Create one
          </Link>
        </Typography>
      </Box>
    </Box>
  )
}
