import { render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { beforeEach, describe, expect, it, vi } from 'vitest';

import { PetInteraction } from './pet-interaction';

const mocks = vi.hoisted(() => ({
  carePet: vi.fn(),
  usePetProfile: vi.fn(),
  usePetInteraction: vi.fn(),
}));

vi.mock('@/entities/gamification-profile', () => ({
  usePetProfile: mocks.usePetProfile,
}));

vi.mock('../../model/use-pet-interaction', () => ({
  usePetInteraction: mocks.usePetInteraction,
}));

describe('PetInteraction', () => {
  beforeEach(() => {
    mocks.carePet.mockReset();
  });

  it('отправляет кормление и поглаживание', async () => {
    const user = userEvent.setup();
    mocks.usePetInteraction.mockReturnValue({
      carePet: mocks.carePet,
      isPending: false,
    });
    mocks.usePetProfile.mockReturnValue({
      data: {
        happiness: 50,
        feedNextAvailableAt: null,
        strokeNextAvailableAt: null,
      },
    });
    render(<PetInteraction />);

    await user.click(screen.getByRole('button', { name: 'Покормить' }));
    await user.click(screen.getByRole('button', { name: 'Погладить' }));

    expect(mocks.carePet).toHaveBeenNthCalledWith(1, 'FEED');
    expect(mocks.carePet).toHaveBeenNthCalledWith(2, 'STROKE');
  });
});
