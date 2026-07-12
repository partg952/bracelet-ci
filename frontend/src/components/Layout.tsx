import { type ReactNode } from 'react'
import { Link, useNavigate } from 'react-router-dom'
import {
  AppBar,
  Toolbar,
  Box,
  Typography,
  Button,
  Avatar,
  Container,
} from '@mui/material'
import { useAuth } from '../context/AuthContext'

export default function Layout({ children }: { children: ReactNode }) {
  const { user, logout } = useAuth()
  const navigate = useNavigate()

  function handleLogout() {
    logout()
    navigate('/login')
  }

  return (
    <Box sx={{ minHeight: '100vh', display: 'flex', flexDirection: 'column', bgcolor: 'background.default' }}>
      <AppBar position="sticky">
        <Toolbar>
          <Box
            component={Link}
            to="/projects"
            sx={{
              display: 'flex',
              alignItems: 'center',
              gap: 1.5,
              textDecoration: 'none',
              mr: 'auto',
            }}
          >
            {/* Wordmark */}
            <Box
              sx={{
                width: 20,
                height: 20,
                bgcolor: '#e2e8f0',
                display: 'flex',
                alignItems: 'center',
                justifyContent: 'center',
                flexShrink: 0,
              }}
            >
              <Box
                component="span"
                sx={{
                  display: 'block',
                  width: 8,
                  height: 8,
                  bgcolor: '#0c0e12',
                }}
              />
            </Box>
            <Typography
              variant="body2"
              sx={{ color: '#e2e8f0', fontWeight: 500, letterSpacing: '0.02em' }}
            >
              braceletci
            </Typography>
          </Box>

          {user && (
            <Box sx={{ display: 'flex', alignItems: 'center', gap: 2 }}>
              <Typography variant="caption" sx={{ color: '#334155' }}>
                {user.email}
              </Typography>
              <Avatar
                sx={{
                  width: 24,
                  height: 24,
                  bgcolor: '#1e2028',
                  color: '#94a3b8',
                  fontSize: '0.7rem',
                  border: '1px solid #334155',
                }}
              >
                {user.username.charAt(0).toUpperCase()}
              </Avatar>
              <Button
                variant="text"
                size="small"
                onClick={handleLogout}
                sx={{ color: '#475569', minWidth: 0, px: 1 }}
              >
                Sign out
              </Button>
            </Box>
          )}
        </Toolbar>
      </AppBar>

      <Container maxWidth="lg" component="main" sx={{ flex: 1, py: 4, px: { xs: 2, sm: 3 } }}>
        {children}
      </Container>
    </Box>
  )
}
