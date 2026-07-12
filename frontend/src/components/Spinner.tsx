import { CircularProgress, Box } from '@mui/material'

interface Props {
  size?: number
  centered?: boolean
}

export default function Spinner({ size = 36, centered = false }: Props) {
  if (centered) {
    return (
      <Box sx={{ display: 'flex', justifyContent: 'center', alignItems: 'center', py: 10 }}>
        <CircularProgress size={size} thickness={3} />
      </Box>
    )
  }
  return <CircularProgress size={size} thickness={3} />
}
