import { act, renderHook, waitFor } from '@testing-library/react';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';

import type { TPet } from '@/entities/gamification-profile';
import { rewardsQueryKeys, type TReward } from '@/entities/reward';
import { createQueryWrapper, createTestQueryClient } from '@/test/render-with-providers';

import { useBuyReward } from './use-buy-reward';

const mocks = vi.hoisted(() => ({
  usePetProfile: vi.fn(),
  updatePetProfile: vi.fn(),
  openChest: vi.fn(),
  addMVPLeaves: vi.fn(),
  toastError: vi.fn(),
}));

vi.mock('@/entities/gamification-profile', () => ({
  usePetProfile: mocks.usePetProfile,
}));

vi.mock('../api/open-chest', () => ({
  openChest: mocks.openChest,
}));

vi.mock('../api/add-mvp-leaves', () => ({
  addMVPLeaves: mocks.addMVPLeaves,
}));

vi.mock('sonner', () => ({
  toast: { error: mocks.toastError },
}));

const pet: TPet = {
  name: 'Листик',
  level: 10,
  leaves: 500,
  nextLevelTargetLeaves: 0,
  chestPrice: 100,
};

const reward: TReward = {
  id: 'reward-1',
  title: 'Бесплатная доставка',
  category: 'FREE_DELIVERY',
  categoryName: 'Доставка',
  source: 'CHEST',
  active: true,
  status: 'ACTIVE',
  expiresAt: '2026-08-20T10:00:00Z',
  awardedAt: '2026-08-10T10:00:00Z',
  redeemedAt: null,
};

const renderBuyRewardHook = (currentPet: TPet | null | undefined = pet) => {
  mocks.usePetProfile.mockReturnValue({
    data: currentPet,
    updatePetProfile: mocks.updatePetProfile,
  });
  const queryClient = createTestQueryClient();

  return {
    queryClient,
    ...renderHook(() => useBuyReward(), {
      wrapper: createQueryWrapper(queryClient),
    }),
  };
};

describe('useBuyReward', () => {
  beforeEach(() => {
    mocks.usePetProfile.mockReset();
    mocks.updatePetProfile.mockReset();
    mocks.openChest.mockReset();
    mocks.addMVPLeaves.mockReset();
    mocks.toastError.mockReset();
  });

  afterEach(() => {
    vi.useRealTimers();
  });

  it.each([
    ['нет профиля', null, true],
    ['не достигнут максимальный уровень', { ...pet, nextLevelTargetLeaves: 1_000 }, true],
    ['недостаточно листьев', { ...pet, leaves: 50 }, true],
    ['все условия выполнены', pet, false],
  ])('определяет доступность сундука: %s', (_caseName, currentPet, expected) => {
    const { result } = renderBuyRewardHook(currentPet);

    expect(result.current.isDisabled).toBe(expected);
  });

  it('открывает награду, показывает её после задержки и обновляет список', async () => {
    vi.useFakeTimers();
    mocks.openChest.mockResolvedValue(reward);
    const { result, queryClient } = renderBuyRewardHook();
    const invalidateQueries = vi.spyOn(queryClient, 'invalidateQueries').mockResolvedValue();

    act(() => result.current.openChest());
    await act(async () => {
      await vi.advanceTimersByTimeAsync(0);
    });

    expect(result.current.reward).toEqual(reward);
    expect(result.current.isOpen).toBe(true);
    expect(result.current.isRewardVisible).toBe(false);
    expect(invalidateQueries).toHaveBeenCalledWith({ queryKey: rewardsQueryKeys.list() });

    await act(async () => {
      await vi.advanceTimersByTimeAsync(1_000);
    });
    expect(result.current.isRewardVisible).toBe(true);
  });

  it('очищает награду и таймер при закрытии', async () => {
    vi.useFakeTimers();
    mocks.openChest.mockResolvedValue(reward);
    const { result } = renderBuyRewardHook();

    act(() => result.current.openChest());
    await act(async () => {
      await vi.advanceTimersByTimeAsync(0);
    });
    act(() => result.current.closeModal());

    expect(result.current.reward).toBeNull();
    expect(result.current.isOpen).toBe(false);
    expect(result.current.isRewardVisible).toBe(false);

    await act(async () => {
      await vi.advanceTimersByTimeAsync(1_000);
    });
    expect(result.current.isRewardVisible).toBe(false);
  });

  it('показывает уведомление при ошибке открытия сундука', async () => {
    mocks.openChest.mockRejectedValue(new Error('Ошибка'));
    const { result } = renderBuyRewardHook();

    act(() => result.current.openChest());

    await waitFor(() => {
      expect(mocks.toastError).toHaveBeenCalledWith('Не удалось открыть сундук');
    });
  });

  it('обновляет профиль после MVP-начисления листьев', async () => {
    const updatedPet = { ...pet, leaves: 700 };
    mocks.addMVPLeaves.mockResolvedValue(updatedPet);
    const { result } = renderBuyRewardHook();

    act(() => result.current.addMVPLeaves());

    await waitFor(() => {
      expect(mocks.updatePetProfile).toHaveBeenCalledWith(updatedPet);
    });
  });

  it('показывает уведомление при ошибке начисления листьев', async () => {
    mocks.addMVPLeaves.mockRejectedValue(new Error('Ошибка'));
    const { result } = renderBuyRewardHook();

    act(() => result.current.addMVPLeaves());

    await waitFor(() => {
      expect(mocks.toastError).toHaveBeenCalledWith('Не удалось начислить листья');
    });
  });
});
