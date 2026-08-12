import { act, renderHook, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';

import { createQueryWrapper, createTestQueryClient } from '@/test/render-with-providers';

import { dailyTasksQueryKeys } from '../api/daily-tasks-keys';
import type { TTask } from '../api/tasks';

import { useDailyTasks } from './use-daily-tasks';

const mocks = vi.hoisted(() => ({
  getTasks: vi.fn(),
  receiveTaskReward: vi.fn(),
  toastError: vi.fn(),
}));

vi.mock('../api/tasks', () => ({
  getTasks: mocks.getTasks,
  receiveTaskReward: mocks.receiveTaskReward,
}));

vi.mock('sonner', () => ({
  toast: { error: mocks.toastError },
}));

const task: TTask = {
  taskId: 'task-1',
  slot: 1,
  type: 'VIEW_LISTINGS',
  description: 'Посмотреть объявления',
  currentCount: 5,
  targetCount: 5,
  rewardLeaves: 20,
  requiredLevel: 1,
  status: 'COMPLETED',
};

describe('useDailyTasks', () => {
  beforeEach(() => {
    mocks.getTasks.mockReset().mockResolvedValue([task]);
    mocks.receiveTaskReward.mockReset();
    mocks.toastError.mockReset();
  });

  it('получает награду и инвалидирует список заданий', async () => {
    mocks.receiveTaskReward.mockResolvedValue({ tasks: [{ ...task, status: 'CLAIMED' }] });
    const queryClient = createTestQueryClient();
    const invalidateQueries = vi.spyOn(queryClient, 'invalidateQueries');
    const { result } = renderHook(() => useDailyTasks(), {
      wrapper: createQueryWrapper(queryClient),
    });
    await waitFor(() => expect(result.current.isSuccess).toBe(true));

    act(() => result.current.receiveReward({ taskId: 'task-1' }));

    await waitFor(() => expect(mocks.receiveTaskReward).toHaveBeenCalledWith('task-1'));
    await waitFor(() =>
      expect(invalidateQueries).toHaveBeenCalledWith({ queryKey: dailyTasksQueryKeys.list() }),
    );
  });

  it('откатывает optimistic update и показывает ошибку', async () => {
    let rejectRequest: (reason?: unknown) => void = () => undefined;
    mocks.receiveTaskReward.mockImplementation(
      () =>
        new Promise((_resolve, reject) => {
          rejectRequest = reject;
        }),
    );
    const queryClient = createTestQueryClient();
    queryClient.setQueryData(dailyTasksQueryKeys.list(), [task]);
    const { result } = renderHook(() => useDailyTasks(), {
      wrapper: createQueryWrapper(queryClient),
    });

    act(() => result.current.receiveReward({ taskId: 'task-1' }));
    await waitFor(() => {
      const cached = queryClient.getQueryData<TTask[]>(dailyTasksQueryKeys.list());
      expect(cached?.[0].status).toBe('CLAIMED');
    });

    act(() => rejectRequest(new Error('Ошибка')));

    await waitFor(() => {
      const cached = queryClient.getQueryData<TTask[]>(dailyTasksQueryKeys.list());
      expect(cached?.[0].status).toBe('COMPLETED');
    });
    expect(mocks.toastError).toHaveBeenCalledWith(
      'Произошла ошибка при получении награды за задание',
    );
  });
});
