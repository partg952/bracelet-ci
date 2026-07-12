import { Chip } from '@mui/material'

interface Props {
  status: string | null | undefined
}

const configs: Record<string, { label: string; color: string; bg: string; border: string }> = {
  passed:    { label: 'Passed',    color: '#4ade80', bg: '#0f1c14', border: '#14532d' },
  failed:    { label: 'Failed',    color: '#f87171', bg: '#1c0f0f', border: '#3f1515' },
  running:   { label: 'Running',   color: '#60a5fa', bg: '#0f1525', border: '#1e3a5f' },
  pending:   { label: 'Pending',   color: '#fb923c', bg: '#1c130f', border: '#431c07' },
  cancelled: { label: 'Cancelled', color: '#64748b', bg: '#111318', border: '#1e2028' },
}

export default function JobStatusBadge({ status }: Props) {
  const key = (status ?? 'pending').toLowerCase()
  const cfg = configs[key] ?? configs['pending']

  return (
    <Chip
      label={cfg.label}
      size="small"
      sx={{
        color: cfg.color,
        backgroundColor: cfg.bg,
        border: `1px solid ${cfg.border}`,
        borderRadius: '2px',
        fontSize: '0.68rem',
        height: 18,
        fontWeight: 400,
        letterSpacing: '0.02em',
        '& .MuiChip-label': { px: '6px' },
      }}
    />
  )
}
