import { render, screen } from '@testing-library/react';
import { describe, expect, it, vi } from 'vitest';

import { PetName } from './pet-name';

const mocks = vi.hoisted(() => ({
  usePetProfile: vi.fn(),
}));

vi.mock('@/entities/gamification-profile', () => ({
  usePetProfile: mocks.usePetProfile,
}));

describe('PetName', () => {
  it.each([
    [34, 'плохое (34.0)'],
    [35, 'нейтральное (35.0)'],
    [79, 'нейтральное (79.0)'],
    [80, 'хорошее (80.0)'],
  ])('показывает статус для настроения %s', (happiness, statusText) => {
    mocks.usePetProfile.mockReturnValue({
      data: {
        name: 'Листик',
        happiness,
      },
    });

    render(<PetName />);

    expect(screen.getByRole('heading', { name: /Листик/ })).toBeInTheDocument();
    expect(screen.getByText(statusText)).toBeInTheDocument();
  });
});
