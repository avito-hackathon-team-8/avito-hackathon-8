import { act, renderHook, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';

import type { TPet } from '@/entities/gamification-profile';
import { createQueryWrapper } from '@/test/render-with-providers';

import { usePetInteraction } from './use-pet-interaction';

const mocks = vi.hoisted(() => ({
  carePetPost: vi.fn(),
  updatePetProfile: vi.fn(),
  usePetProfile: vi.fn(),
}));

vi.mock('@/entities/gamification-profile', () => ({
  usePetProfile: mocks.usePetProfile,
}));

vi.mock('../api/pet-interactions', () => ({
  carePetPost: mocks.carePetPost,
}));

const pet: TPet = {
  name: 'Листик',
  level: 3,
  leaves: 250,
  nextLevelTargetLeaves: 500,
  chestPrice: 100,
  bowlImageUrl: null,
  bedImageUrl: null,
  happiness: 40,
  happinessMultiplier: 0.9,
  calculatedAt: '2026-08-12T12:52:25.179950567Z',
  decaysToZeroAt: '2026-08-15T12:52:15.223227999Z',
  feedNextAvailableAt: null,
  strokeNextAvailableAt: null,
};

describe('usePetInteraction', () => {
  beforeEach(() => {
    mocks.carePetPost.mockReset();
    mocks.updatePetProfile.mockReset();
    mocks.usePetProfile.mockReset().mockReturnValue({
      data: pet,
      updatePetProfile: mocks.updatePetProfile,
    });
  });

  it.each(['FEED', 'STROKE'] as const)(
    'отправляет действие %s и обновляет состояние питомца',
    async (type) => {
      const petState = {
        happiness: 100,
        happinessMultiplier: 1.5,
        calculatedAt: '2026-08-12T12:19:51.697Z',
        decaysToZeroAt: '2026-08-15T12:19:51.697Z',
        strokeNextAvailableAt: '2026-08-12T18:19:51.697Z',
        feedNextAvailableAt: '2026-08-12T18:19:51.697Z',
      };
      mocks.carePetPost.mockResolvedValue(petState);
      const { result } = renderHook(() => usePetInteraction(), {
        wrapper: createQueryWrapper(),
      });

      act(() => result.current.carePet(type));

      await waitFor(() => expect(mocks.carePetPost).toHaveBeenCalledWith({ type }));
      await waitFor(() => {
        expect(mocks.updatePetProfile).toHaveBeenCalledWith({
          ...pet,
          ...petState,
        });
      });
    },
  );
});
