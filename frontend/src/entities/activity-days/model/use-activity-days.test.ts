import { act, renderHook, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';

import { createQueryWrapper, createTestQueryClient } from '@/test/render-with-providers';

import type { TResponseActivityDay } from '../api/activity-day';
import { activityDayKeys } from '../api/activity-days-keys';

import { useActivityDays } from './use-activity-days';

const mocks = vi.hoisted(() => ({
  getActivityDay: vi.fn(),
  receiveActivityDayReward: vi.fn(),
  toastError: vi.fn(),
}));

vi.mock('../api/activity-day', () => ({
  getActivityDay: mocks.getActivityDay,
  receiveActivityDayReward: mocks.receiveActivityDayReward,
}));

vi.mock('sonner', () => ({
  toast: { error: mocks.toastError },
}));

const week: TResponseActivityDay = {
  claimedDaysCount: 2,
  claims: [
    {
      weekday: 1,
      date: '2026-08-10',
      status: 'CLAIMED',
      rewardLeaves: 10,
      baseRewardLeaves: 10,
      claimId: 'claim-1',
    },
    {
      weekday: 2,
      date: '2026-08-11',
      status: 'AVAILABLE',
      rewardLeaves: 20,
      baseRewardLeaves: 20,
      claimId: 'claim-2',
    },
  ],
};

describe('useActivityDays', () => {
  beforeEach(() => {
    mocks.getActivityDay.mockReset().mockResolvedValue(week);
    mocks.receiveActivityDayReward.mockReset();
    mocks.toastError.mockReset();
  });

  it('загружает данные недели по умолчанию', async () => {
    const queryClient = createTestQueryClient();
    const { result } = renderHook(() => useActivityDays(), {
      wrapper: createQueryWrapper(queryClient),
    });

    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(mocks.getActivityDay).toHaveBeenCalledOnce();
    expect(result.current.data).toEqual(week);
  });

  it('не выполняет запрос при enabled=false', () => {
    const queryClient = createTestQueryClient();
    const { result } = renderHook(() => useActivityDays({ enabled: false }), {
      wrapper: createQueryWrapper(queryClient),
    });

    expect(result.current.fetchStatus).toBe('idle');
    expect(mocks.getActivityDay).not.toHaveBeenCalled();
  });

  it('после получения награды инвалидирует и обновляет данные недели', async () => {
    mocks.receiveActivityDayReward.mockResolvedValue(undefined);
    const queryClient = createTestQueryClient();
    const invalidateQueries = vi.spyOn(queryClient, 'invalidateQueries');
    const fetchQuery = vi.spyOn(queryClient, 'fetchQuery').mockResolvedValue({
      ...week,
      claimedDaysCount: 3,
    });
    const { result } = renderHook(() => useActivityDays(), {
      wrapper: createQueryWrapper(queryClient),
    });
    await waitFor(() => expect(result.current.isSuccess).toBe(true));

    act(() => result.current.receiveReward());

    await waitFor(() => expect(mocks.receiveActivityDayReward).toHaveBeenCalledOnce());
    await waitFor(() => {
      expect(invalidateQueries).toHaveBeenCalledWith({
        queryKey: activityDayKeys.week(),
        refetchType: 'none',
      });
      expect(fetchQuery).toHaveBeenCalledWith({
        queryKey: activityDayKeys.week(),
        queryFn: mocks.getActivityDay,
      });
    });
  });

  it('показывает ошибку получения награды', async () => {
    mocks.receiveActivityDayReward.mockRejectedValue(new Error('Ошибка'));
    const queryClient = createTestQueryClient();
    const { result } = renderHook(() => useActivityDays(), {
      wrapper: createQueryWrapper(queryClient),
    });
    await waitFor(() => expect(result.current.isSuccess).toBe(true));

    act(() => result.current.receiveReward());

    await waitFor(() =>
      expect(mocks.toastError).toHaveBeenCalledWith(
        'Произошла ошибка при получении ежедневной награды',
      ),
    );
  });

  it('сообщает об ошибке повторного получения недели', async () => {
    mocks.receiveActivityDayReward.mockResolvedValue(undefined);
    const queryClient = createTestQueryClient();
    vi.spyOn(queryClient, 'fetchQuery').mockRejectedValue(new Error('Ошибка обновления'));
    const { result } = renderHook(() => useActivityDays(), {
      wrapper: createQueryWrapper(queryClient),
    });
    await waitFor(() => expect(result.current.isSuccess).toBe(true));

    act(() => result.current.receiveReward());

    await waitFor(() =>
      expect(mocks.toastError).toHaveBeenCalledWith('Не удалось обновить данные за неделю'),
    );
  });
});
