import { useNavigate } from 'react-router-dom'
import {
  Box,
  Button,
  Container,
  Typography,
  Divider,
} from '@mui/material'
import logo from '../assets/logo_without_title.png'
import logoWithTitle from '../assets/logo.png'

const features = [
  {
    title: 'Push to trigger',
    body: 'Webhook-driven pipelines start the moment you push. No manual steps, no waiting.',
  },
  {
    title: 'Live job stream',
    body: 'Watch your builds update in real time over SSE — no polling, no refreshing.',
  },
  {
    title: 'Container isolation',
    body: 'Every run executes inside its own Docker container. Clean state, every time.',
  },
  {
    title: 'Simple data model',
    body: 'Projects, jobs, users. Nothing extra. Easy to reason about, easy to extend.',
  },
]

export default function LandingPage() {
  const navigate = useNavigate()

  return (
    <Box sx={{ bgcolor: 'background.default', minHeight: '100vh', display: 'flex', flexDirection: 'column' }}>

      {/* ── Nav ── */}
      <Box
        component="header"
        sx={{
          borderBottom: '1px solid #1e2028',
          px: { xs: 3, md: 6 },
          py: 1.5,
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'space-between',
        }}
      >
        <Box component="img" src={logo} alt="BraceletCI" sx={{ height: 28 }} />

        <Box sx={{ display: 'flex', gap: 1 }}>
          <Button
            variant="text"
            size="small"
            onClick={() => navigate('/login')}
            sx={{ color: '#64748b' }}
          >
            Sign in
          </Button>
          <Button
            variant="contained"
            size="small"
            onClick={() => navigate('/register')}
          >
            Get started
          </Button>
        </Box>
      </Box>

      {/* ── Hero ── */}
      <Container maxWidth="md" sx={{ flex: 1, display: 'flex', flexDirection: 'column', justifyContent: 'center', pt: { xs: 8, md: 12 }, pb: 6, textAlign: 'center' }}>

        <Box sx={{ mb: 4, display: 'flex', justifyContent: 'center' }}>
          <Box component="img" src={logoWithTitle} alt="BraceletCI" sx={{ height: 160, width: 'auto' }} />
        </Box>

        <Typography
          variant="h1"
          sx={{
            fontSize: { xs: '2rem', md: '2.75rem' },
            fontWeight: 600,
            color: '#e2e8f0',
            letterSpacing: '-0.03em',
            lineHeight: 1.15,
            mb: 2,
          }}
        >
          CI that stays out of your way
        </Typography>

        <Typography
          variant="body1"
          sx={{
            color: '#475569',
            fontSize: '1rem',
            maxWidth: 480,
            mx: 'auto',
            mb: 5,
            lineHeight: 1.7,
          }}
        >
          Run your tests on every push — just a webhook, a Docker container, and results.
        </Typography>

        <Box sx={{ display: 'flex', gap: 1.5, justifyContent: 'center', flexWrap: 'wrap' }}>
          <Button
            variant="contained"
            size="large"
            onClick={() => navigate('/register')}
            sx={{ px: 3, py: '9px', fontSize: '0.85rem' }}
          >
            Create an account
          </Button>
          <Button
            variant="outlined"
            size="large"
            onClick={() => navigate('/login')}
            sx={{ px: 3, py: '9px', fontSize: '0.85rem', borderColor: '#1e2028', color: '#64748b' }}
          >
            Sign in
          </Button>
        </Box>
      </Container>

      {/* ── Features ── */}
      <Box sx={{ borderTop: '1px solid #1e2028', py: { xs: 6, md: 8 } }}>
        <Container maxWidth="lg">
          <Box
            sx={{
              display: 'grid',
              gridTemplateColumns: { xs: '1fr', sm: '1fr 1fr', md: '1fr 1fr 1fr 1fr' },
              gap: 0,
            }}
          >
            {features.map((f, i) => (
              <Box
                key={f.title}
                sx={{
                  px: 3,
                  py: 3,
                  borderLeft: i > 0 ? { sm: i % 2 === 0 ? 'none' : '1px solid #1e2028', md: '1px solid #1e2028' } : 'none',
                  borderTop: { xs: i > 0 ? '1px solid #1e2028' : 'none', sm: i >= 2 ? '1px solid #1e2028' : 'none', md: 'none' },
                }}
              >
                <Typography
                  variant="subtitle2"
                  sx={{ color: '#e2e8f0', mb: 1, fontWeight: 500 }}
                >
                  {f.title}
                </Typography>
                <Typography variant="body2" sx={{ color: '#475569', lineHeight: 1.65 }}>
                  {f.body}
                </Typography>
              </Box>
            ))}
          </Box>
        </Container>
      </Box>

      {/* ── Footer ── */}
      <Divider />
      <Box sx={{ py: 2.5, px: 4, display: 'flex', alignItems: 'center', justifyContent: 'space-between', flexWrap: 'wrap', gap: 1 }}>
        <Box component="img" src={logo} alt="BraceletCI" sx={{ height: 18, opacity: 0.4 }} />
        <Typography variant="caption" sx={{ color: '#1e2028' }}>
          open source · MIT
        </Typography>
      </Box>
    </Box>
  )
}
