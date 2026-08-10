import { act, renderHook, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';

import { rewardsQueryKeys } from '@/entities/reward';
import { createQueryWrapper, createTestQueryClient } from '@/test/render-with-providers';

import { gamificationProfileKeys } from '../api/gamification-profile-keys';
import type { TLevelRewardItem } from '../api/levels-rewards';

import { useLevelsProfile } from './use-levels-profile';

const mocks = vi.hoisted(() => ({
  getLevelsRewards: vi.fn(),
  receiveLevelReward: vi.fn(),
  toastError: vi.fn(),
}));

vi.mock('../api/levels-rewards', () => ({
  getLevelsRewards: mocks.getLevelsRewards,
  receiveLevelReward: mocks.receiveLevelReward,
}));

vi.mock('sonner', () => ({
  toast: { error: mocks.toastError },
}));

const level: TLevelRewardItem = {
  level: 3,
  status: 'UNOPENED',
  reward: {
    id: 'reward-3',
    type: 'FREE_DELIVERY',
    description: 'Бесплатная доставка',
  },
  expiresAt: '2026-08-20T10:00:00Z',
};

describe('useLevelsProfile', () => {
  beforeEach(() => {
    mocks.getLevelsRewards.mockReset().mockResolvedValue({ levels: [level] });
    mocks.receiveLevelReward.mockReset();
    mocks.toastError.mockReset();
  });

  it('получает награду и обновляет уровни и список наград', async () => {
    mocks.receiveLevelReward.mockResolvedValue({ levels: [{ level: 3, status: 'CLAIMED' }] });
    const queryClient = createTestQueryClient();
    const invalidateQueries = vi.spyOn(queryClient, 'invalidateQueries');
    const { result } = renderHook(() => useLevelsProfile(), {
      wrapper: createQueryWrapper(queryClient),
    });
    await waitFor(() => expect(result.current.isSuccess).toBe(true));

    act(() => result.current.receiveReward({ rewardId: 'reward-3', level: 3 }));

    await waitFor(() => expect(mocks.receiveLevelReward).toHaveBeenCalledWith('reward-3'));
    await waitFor(() => {
      expect(invalidateQueries).toHaveBeenCalledWith({ queryKey: rewardsQueryKeys.list() });
      expect(invalidateQueries).toHaveBeenCalledWith({
        queryKey: gamificationProfileKeys.levels(),
      });
    });
  });

  it('откатывает optimistic update при ошибке', async () => {
    let rejectRequest: (reason?: unknown) => void = () => undefined;
    mocks.receiveLevelReward.mockImplementation(
      () =>
        new Promise((_resolve, reject) => {
          rejectRequest = reject;
        }),
    );
    const queryClient = createTestQueryClient();
    queryClient.setQueryData(gamificationProfileKeys.levels(), { levels: [level] });
    const { result } = renderHook(() => useLevelsProfile(), {
      wrapper: createQueryWrapper(queryClient),
    });

    act(() => result.current.receiveReward({ rewardId: 'reward-3', level: 3 }));
    await waitFor(() => {
      const cached = queryClient.getQueryData<{ levels: TLevelRewardItem[] }>(
        gamificationProfileKeys.levels(),
      );
      expect(cached?.levels[0].status).toBe('CLAIMED');
    });

    act(() => rejectRequest(new Error('Ошибка')));

    await waitFor(() => {
      const cached = queryClient.getQueryData<{ levels: TLevelRewardItem[] }>(
        gamificationProfileKeys.levels(),
      );
      expect(cached?.levels[0].status).toBe('UNOPENED');
    });
    expect(mocks.toastError).toHaveBeenCalledWith(
      'Произошла ошибка при получении награды за уровень',
    );
  });
});
