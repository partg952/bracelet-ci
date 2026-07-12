import { createTheme } from '@mui/material/styles'

const theme = createTheme({
  palette: {
    mode: 'dark',
    primary: {
      main: '#e2e8f0',
      light: '#f8fafc',
      dark: '#94a3b8',
      contrastText: '#0f172a',
    },
    secondary: {
      main: '#94a3b8',
    },
    error: {
      main: '#f87171',
    },
    warning: {
      main: '#fb923c',
    },
    success: {
      main: '#4ade80',
    },
    info: {
      main: '#60a5fa',
    },
    background: {
      default: '#0c0e12',
      paper: '#111318',
    },
    divider: '#1e2028',
    text: {
      primary: '#e2e8f0',
      secondary: '#64748b',
      disabled: '#334155',
    },
  },
  typography: {
    fontFamily: '"Inter", "SF Pro Display", system-ui, sans-serif',
    fontSize: 13,
    h4: { fontSize: '1.25rem', fontWeight: 500, letterSpacing: '-0.01em' },
    h5: { fontSize: '1.05rem', fontWeight: 500, letterSpacing: '-0.01em' },
    h6: { fontSize: '0.9rem', fontWeight: 500 },
    subtitle1: { fontSize: '0.875rem', fontWeight: 500 },
    subtitle2: { fontSize: '0.8rem', fontWeight: 500 },
    body1: { fontSize: '0.875rem' },
    body2: { fontSize: '0.8rem' },
    caption: { fontSize: '0.7rem', letterSpacing: '0.03em' },
    button: { textTransform: 'none', fontWeight: 400, fontSize: '0.8rem', letterSpacing: '0.01em' },
  },
  shape: {
    borderRadius: 2,
  },
  components: {
    MuiCssBaseline: {
      styleOverrides: {
        body: { backgroundColor: '#0c0e12' },
      },
    },
    MuiButton: {
      defaultProps: { disableElevation: true, disableRipple: false },
      styleOverrides: {
        root: {
          borderRadius: 2,
          padding: '5px 14px',
          fontSize: '0.8rem',
        },
        contained: {
          backgroundColor: '#e2e8f0',
          color: '#0f172a',
          border: '1px solid transparent',
          '&:hover': { backgroundColor: '#f8fafc' },
          '&.Mui-disabled': { backgroundColor: '#1e2028', color: '#334155' },
        },
        outlined: {
          borderColor: '#1e2028',
          color: '#94a3b8',
          '&:hover': { borderColor: '#334155', backgroundColor: '#111318' },
        },
        text: {
          color: '#94a3b8',
          '&:hover': { backgroundColor: '#111318', color: '#e2e8f0' },
        },
      },
    },
    MuiTextField: {
      defaultProps: { variant: 'outlined', size: 'small' },
      styleOverrides: {
        root: {
          '& .MuiOutlinedInput-root': {
            fontSize: '0.8rem',
            borderRadius: 2,
            backgroundColor: '#0c0e12',
            '& fieldset': { borderColor: '#1e2028' },
            '&:hover fieldset': { borderColor: '#334155' },
            '&.Mui-focused fieldset': { borderColor: '#475569', borderWidth: 1 },
          },
          '& .MuiInputLabel-root': { fontSize: '0.8rem', color: '#475569' },
          '& .MuiInputLabel-root.Mui-focused': { color: '#94a3b8' },
          '& input': { color: '#e2e8f0' },
        },
      },
    },
    MuiCard: {
      defaultProps: { elevation: 0 },
      styleOverrides: {
        root: {
          backgroundColor: '#111318',
          backgroundImage: 'none',
          border: '1px solid #1e2028',
          borderRadius: 2,
        },
      },
    },
    MuiPaper: {
      defaultProps: { elevation: 0 },
      styleOverrides: {
        root: {
          backgroundImage: 'none',
          backgroundColor: '#111318',
          border: '1px solid #1e2028',
          borderRadius: 2,
        },
      },
    },
    MuiAppBar: {
      styleOverrides: {
        root: {
          backgroundColor: '#0c0e12',
          backgroundImage: 'none',
          borderBottom: '1px solid #1e2028',
          boxShadow: 'none',
        },
      },
    },
    MuiToolbar: {
      styleOverrides: {
        root: { minHeight: '48px !important' },
      },
    },
    MuiChip: {
      styleOverrides: {
        root: {
          borderRadius: 2,
          fontSize: '0.7rem',
          height: 20,
          fontWeight: 400,
        },
        sizeSmall: { height: 20 },
      },
    },
    MuiTableCell: {
      styleOverrides: {
        root: {
          borderColor: '#1e2028',
          padding: '10px 16px',
          fontSize: '0.8rem',
        },
        head: {
          color: '#475569',
          fontWeight: 400,
          fontSize: '0.7rem',
          textTransform: 'uppercase',
          letterSpacing: '0.06em',
          backgroundColor: '#0c0e12',
          borderBottom: '1px solid #1e2028',
        },
      },
    },
    MuiTableRow: {
      styleOverrides: {
        root: {
          '&:hover td': { backgroundColor: '#14161c' },
          '&:last-child td': { border: 0 },
          transition: 'none',
        },
      },
    },
    MuiDivider: {
      styleOverrides: { root: { borderColor: '#1e2028' } },
    },
    MuiAlert: {
      styleOverrides: {
        root: { borderRadius: 2, fontSize: '0.8rem' },
      },
    },
    MuiTooltip: {
      styleOverrides: {
        tooltip: {
          backgroundColor: '#1e2028',
          border: '1px solid #334155',
          fontSize: '0.72rem',
          borderRadius: 2,
          color: '#cbd5e1',
        },
        arrow: { color: '#1e2028' },
      },
    },
    MuiSkeleton: {
      styleOverrides: {
        root: { backgroundColor: '#14161c', borderRadius: 2 },
      },
    },
    MuiAvatar: {
      styleOverrides: {
        root: { borderRadius: 2, fontSize: '0.75rem', fontWeight: 500 },
      },
    },
    MuiIconButton: {
      styleOverrides: {
        root: {
          borderRadius: 2,
          '&:hover': { backgroundColor: '#1e2028' },
        },
      },
    },
  },
})

export default theme
