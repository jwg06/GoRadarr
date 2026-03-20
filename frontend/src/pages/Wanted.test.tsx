import { screen } from '@testing-library/react'
import Wanted from './Wanted'
import { renderWithProviders } from '../test/test-utils'
import api from '../lib/api'

vi.mock('../lib/api', () => ({
  default: { get: vi.fn(), post: vi.fn() },
}))

const mockedApi = api as unknown as { get: ReturnType<typeof vi.fn>; post?: ReturnType<typeof vi.fn>; delete?: ReturnType<typeof vi.fn>; put?: ReturnType<typeof vi.fn> }

describe('Wanted', () => {
  it('filters down to monitored missing movies', async () => {
    mockedApi.get.mockResolvedValue({
      data: [
        { id: 1, title: 'Dune', sortTitle: 'Dune', tmdbId: 1, status: 'released', year: 2021, runtime: 150, qualityProfileId: 1, monitored: true, minimumAvailability: 'released', hasFile: false, added: new Date().toISOString() },
        { id: 2, title: 'Alien', sortTitle: 'Alien', tmdbId: 2, status: 'released', year: 1979, runtime: 117, qualityProfileId: 1, monitored: true, minimumAvailability: 'released', hasFile: true, added: new Date().toISOString() },
      ],
    })

    renderWithProviders(<Wanted />)

    expect(await screen.findByText('Dune')).toBeInTheDocument()
    expect(screen.queryByText('Alien')).not.toBeInTheDocument()
  })
})
