import { act, renderHook, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';

import { createQueryWrapper, createTestQueryClient } from '@/test/render-with-providers';

import { gamificationProfileKeys } from '../api/gamification-profile-keys';

import { usePetProfile } from './use-pet-profile';

const mocks = vi.hoisted(() => ({
  getPetName: vi.fn(),
}));

vi.mock('../api/pet', () => ({
  getPetName: mocks.getPetName,
}));

const pet = {
  name: 'Листик',
  level: 3,
  leaves: 250,
  nextLevelTargetLeaves: 500,
  chestPrice: 100,
};

describe('usePetProfile', () => {
  beforeEach(() => {
    mocks.getPetName.mockReset().mockResolvedValue(pet);
  });

  it('загружает профиль питомца', async () => {
    const queryClient = createTestQueryClient();
    const { result } = renderHook(() => usePetProfile(), {
      wrapper: createQueryWrapper(queryClient),
    });

    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(result.current.data).toEqual(pet);
  });

  it('позволяет синхронно обновить профиль в кэше', async () => {
    const queryClient = createTestQueryClient();
    const { result } = renderHook(() => usePetProfile(), {
      wrapper: createQueryWrapper(queryClient),
    });
    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    const updatedPet = { ...pet, leaves: 300 };

    act(() => result.current.updatePetProfile(updatedPet));

    expect(queryClient.getQueryData(gamificationProfileKeys.pet())).toEqual(updatedPet);
    await waitFor(() => expect(result.current.data).toEqual(updatedPet));
  });
});
