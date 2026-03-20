import { screen } from '@testing-library/react'
import Queue from './Queue'
import { renderWithProviders } from '../test/test-utils'
import api from '../lib/api'

vi.mock('../lib/api', () => ({
  default: { get: vi.fn(), delete: vi.fn(), put: vi.fn() },
}))

const mockedApi = api as unknown as { get: ReturnType<typeof vi.fn>; post?: ReturnType<typeof vi.fn>; delete?: ReturnType<typeof vi.fn>; put?: ReturnType<typeof vi.fn> }

describe('Queue', () => {
  it('renders queue items and status cards', async () => {
    mockedApi.get.mockImplementation((url: string) => {
      if (url === '/queue/status') {
        return Promise.resolve({ data: { totalCount: 1, count: 1, warnings: false, errors: false } })
      }
      return Promise.resolve({ data: { page: 1, pageSize: 50, totalRecords: 1, records: [{ id: 7, movieId: 2, title: 'Interstellar', size: 100, sizeleft: 40, status: 'downloading', trackedDownloadStatus: 'ok', trackedDownloadState: 'downloading', downloadId: 'abc', protocol: 'torrent', downloadClient: 'qBittorrent' }] } })
    })

    renderWithProviders(<Queue />)

    expect(await screen.findByText('Interstellar')).toBeInTheDocument()
    expect(screen.getByText('Total queue items')).toBeInTheDocument()
  })
})
