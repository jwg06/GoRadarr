import { fireEvent, screen } from '@testing-library/react'
import Activity from './Activity'
import { renderWithProviders } from '../test/test-utils'
import api from '../lib/api'

vi.mock('../lib/api', () => ({
  default: { get: vi.fn() },
}))

const mockedApi = api as unknown as { get: ReturnType<typeof vi.fn> }

const mockLogs = [
  { time: '2024-01-01T10:00:00Z', level: 'INFO', msg: 'Server started' },
  { time: '2024-01-01T10:01:00Z', level: 'WARN', msg: 'Connection slow' },
  { time: '2024-01-01T10:02:00Z', level: 'ERROR', msg: 'Request failed' },
  { time: '2024-01-01T10:03:00Z', level: 'DEBUG', msg: 'Debugging output' },
]

describe('Activity', () => {
  beforeEach(() => {
    mockedApi.get.mockImplementation((url: string) => {
      if (url === '/log') return Promise.resolve({ data: mockLogs })
      return Promise.resolve({ data: [] })
    })
  })

  afterEach(() => {
    vi.clearAllMocks()
  })

  it('renders the Logs tab with all log entries', async () => {
    renderWithProviders(<Activity />)

    expect(await screen.findByText('Server started')).toBeInTheDocument()
    expect(screen.getByText('Connection slow')).toBeInTheDocument()
    expect(screen.getByText('Request failed')).toBeInTheDocument()
    expect(screen.getByText('Debugging output')).toBeInTheDocument()
  })

  it('shows level badges for each log entry', async () => {
    renderWithProviders(<Activity />)

    await screen.findByText('Server started')

    // Badges in the table rows
    const badges = screen.getAllByText('INFO')
    expect(badges.length).toBeGreaterThanOrEqual(1)
  })

  it('level filter hides non-matching rows', async () => {
    renderWithProviders(<Activity />)

    // Wait for initial load
    expect(await screen.findByText('Server started')).toBeInTheDocument()

    // Click the ERROR filter pill
    fireEvent.click(screen.getByRole('button', { name: 'ERROR' }))

    // Only ERROR entries should remain
    expect(screen.queryByText('Server started')).not.toBeInTheDocument()
    expect(screen.queryByText('Connection slow')).not.toBeInTheDocument()
    expect(screen.queryByText('Debugging output')).not.toBeInTheDocument()
    expect(screen.getByText('Request failed')).toBeInTheDocument()
  })

  it('ALL filter restores all rows', async () => {
    renderWithProviders(<Activity />)

    await screen.findByText('Server started')

    // Filter to INFO
    fireEvent.click(screen.getByRole('button', { name: 'INFO' }))
    expect(screen.queryByText('Request failed')).not.toBeInTheDocument()

    // Restore to ALL
    fireEvent.click(screen.getByRole('button', { name: 'ALL' }))
    expect(screen.getByText('Request failed')).toBeInTheDocument()
    expect(screen.getByText('Connection slow')).toBeInTheDocument()
  })

  it('renders Logs and Tasks tabs', async () => {
    renderWithProviders(<Activity />)

    expect(screen.getByRole('button', { name: 'Logs' })).toBeInTheDocument()
    expect(screen.getByRole('button', { name: 'Tasks' })).toBeInTheDocument()
  })
})
