import { screen } from '@testing-library/react'
import Sidebar from './Sidebar'
import { renderWithProviders } from '../test/test-utils'

describe('Sidebar', () => {
  it('renders advanced navigation items', () => {
    renderWithProviders(<Sidebar />)

    expect(screen.getByText('Queue')).toBeInTheDocument()
    expect(screen.getByText('Wanted')).toBeInTheDocument()
    expect(screen.getByText('Settings')).toBeInTheDocument()
  })
})
