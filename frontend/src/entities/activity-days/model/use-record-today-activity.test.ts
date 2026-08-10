import { renderHook, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';

import { createQueryWrapper, createTestQueryClient } from '@/test/render-with-providers';

import { activityDayKeys } from '../api/activity-days-keys';

import { useRecordTodayActivity } from './use-record-today-activity';

const mocks = vi.hoisted(() => ({
  getActivityDay: vi.fn(),
  recordTodayActivity: vi.fn(),
  toastError: vi.fn(),
}));

vi.mock('../api/activity-day', () => ({
  getActivityDay: mocks.getActivityDay,
  recordTodayActivity: mocks.recordTodayActivity,
}));

vi.mock('sonner', () => ({
  toast: { error: mocks.toastError },
}));

describe('useRecordTodayActivity', () => {
  beforeEach(() => {
    mocks.getActivityDay.mockReset().mockResolvedValue({ claimedDaysCount: 1, claims: [] });
    mocks.recordTodayActivity.mockReset().mockResolvedValue(undefined);
    mocks.toastError.mockReset();
  });

  it('записывает активность только один раз за жизненный цикл хука', async () => {
    const queryClient = createTestQueryClient();
    const { rerender } = renderHook(() => useRecordTodayActivity(), {
      wrapper: createQueryWrapper(queryClient),
    });

    await waitFor(() => expect(mocks.recordTodayActivity).toHaveBeenCalledOnce());
    rerender();
    expect(mocks.recordTodayActivity).toHaveBeenCalledOnce();
  });

  it('после записи инвалидирует и обновляет данные недели', async () => {
    const queryClient = createTestQueryClient();
    const invalidateQueries = vi.spyOn(queryClient, 'invalidateQueries');
    const fetchQuery = vi.spyOn(queryClient, 'fetchQuery').mockResolvedValue({
      claimedDaysCount: 1,
      claims: [],
    });

    renderHook(() => useRecordTodayActivity(), {
      wrapper: createQueryWrapper(queryClient),
    });

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

  it('показывает ошибку записи активности', async () => {
    mocks.recordTodayActivity.mockRejectedValue(new Error('Ошибка'));
    const queryClient = createTestQueryClient();

    renderHook(() => useRecordTodayActivity(), {
      wrapper: createQueryWrapper(queryClient),
    });

    await waitFor(() =>
      expect(mocks.toastError).toHaveBeenCalledWith('Не удалось отметить активность за сегодня'),
    );
  });

  it('показывает ошибку обновления недели после успешной записи', async () => {
    const queryClient = createTestQueryClient();
    vi.spyOn(queryClient, 'fetchQuery').mockRejectedValue(new Error('Ошибка обновления'));

    renderHook(() => useRecordTodayActivity(), {
      wrapper: createQueryWrapper(queryClient),
    });

    await waitFor(() =>
      expect(mocks.toastError).toHaveBeenCalledWith('Не удалось получить данные за неделю'),
    );
  });
});
