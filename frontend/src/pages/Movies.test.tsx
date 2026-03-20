import { screen } from '@testing-library/react'
import Movies from './Movies'
import { renderWithProviders } from '../test/test-utils'
import api from '../lib/api'

vi.mock('../lib/api', () => ({
  default: { get: vi.fn() },
}))

const mockedApi = api as unknown as { get: ReturnType<typeof vi.fn>; post?: ReturnType<typeof vi.fn>; delete?: ReturnType<typeof vi.fn>; put?: ReturnType<typeof vi.fn> }

describe('Movies', () => {
  it('shows movie stats and cards', async () => {
    mockedApi.get.mockImplementation((url: string) => {
      if (url === '/movie') {
        return Promise.resolve({ data: [{ id: 1, title: 'Arrival', sortTitle: 'Arrival', tmdbId: 1, status: 'released', year: 2016, runtime: 116, qualityProfileId: 1, monitored: true, minimumAvailability: 'released', hasFile: false, added: new Date().toISOString() }] })
      }
      return Promise.resolve({ data: [{ id: 1, name: 'HD-1080p', upgradeAllowed: true, cutoff: 7 }] })
    })

    renderWithProviders(<Movies />)

    expect(await screen.findByText('Arrival')).toBeInTheDocument()
    expect(screen.getByText('Total Movies')).toBeInTheDocument()
    expect(screen.getByText('Wanted')).toBeInTheDocument()
  })
})
