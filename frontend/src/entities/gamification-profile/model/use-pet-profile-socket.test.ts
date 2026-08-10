import { act, renderHook } from '@testing-library/react';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';

import { dailyTasksQueryKeys } from '@/entities/daily-task';
import { sessionStorageKeysMap } from '@/shared/lib';
import { MockWebSocket } from '@/test/mock-web-socket';
import { createQueryWrapper, createTestQueryClient } from '@/test/render-with-providers';

import { gamificationProfileKeys } from '../api/gamification-profile-keys';
import type { TPet } from '../api/pet';

import { usePetProfileSocket } from './use-pet-profile-socket';

const mocks = vi.hoisted(() => ({
  updatePetProfile: vi.fn(),
}));

vi.mock('./use-pet-profile', () => ({
  usePetProfile: () => ({ updatePetProfile: mocks.updatePetProfile }),
}));

const pet: TPet = {
  name: 'Листик',
  level: 3,
  leaves: 300,
  nextLevelTargetLeaves: 500,
  chestPrice: 100,
};

describe('usePetProfileSocket', () => {
  beforeEach(() => {
    vi.useFakeTimers();
    vi.stubGlobal('WebSocket', MockWebSocket);
    MockWebSocket.reset();
    mocks.updatePetProfile.mockReset();
    vi.spyOn(console, 'error').mockImplementation(() => undefined);
    vi.spyOn(console, 'warn').mockImplementation(() => undefined);
    vi.spyOn(console, 'log').mockImplementation(() => undefined);
  });

  afterEach(() => {
    vi.useRealTimers();
    vi.unstubAllGlobals();
  });

  it('не подключается, когда синхронизация выключена', async () => {
    sessionStorage.setItem(sessionStorageKeysMap.authToken, 'token-1');
    const queryClient = createTestQueryClient();
    renderHook(() => usePetProfileSocket({ enabled: false }), {
      wrapper: createQueryWrapper(queryClient),
    });

    await act(async () => vi.runOnlyPendingTimersAsync());
    expect(MockWebSocket.instances).toHaveLength(0);
  });

  it('обновляет профиль и связанные данные при повышении уровня', async () => {
    sessionStorage.setItem(sessionStorageKeysMap.authToken, 'token-1');
    const queryClient = createTestQueryClient();
    queryClient.setQueryData(gamificationProfileKeys.pet(), { ...pet, level: 2 });
    const invalidateQueries = vi.spyOn(queryClient, 'invalidateQueries').mockResolvedValue();
    renderHook(() => usePetProfileSocket({ enabled: true }), {
      wrapper: createQueryWrapper(queryClient),
    });
    await act(async () => vi.advanceTimersByTimeAsync(0));

    act(() => {
      MockWebSocket.instances[0].emitMessage({
        event: 'PET_PROGRESS_UPDATED',
        data: pet,
      });
    });

    expect(mocks.updatePetProfile).toHaveBeenCalledWith(pet);
    expect(invalidateQueries).toHaveBeenCalledWith({
      queryKey: gamificationProfileKeys.levels(),
    });
    expect(invalidateQueries).toHaveBeenCalledWith({ queryKey: dailyTasksQueryKeys.list() });
  });

  it('не обновляет связанные запросы без повышения уровня', async () => {
    sessionStorage.setItem(sessionStorageKeysMap.authToken, 'token-1');
    const queryClient = createTestQueryClient();
    queryClient.setQueryData(gamificationProfileKeys.pet(), pet);
    const invalidateQueries = vi.spyOn(queryClient, 'invalidateQueries').mockResolvedValue();
    renderHook(() => usePetProfileSocket({ enabled: true }), {
      wrapper: createQueryWrapper(queryClient),
    });
    await act(async () => vi.advanceTimersByTimeAsync(0));

    act(() => {
      MockWebSocket.instances[0].emitMessage({
        event: 'PET_PROGRESS_UPDATED',
        data: { ...pet, leaves: 350, levelUp: false },
      });
    });

    expect(mocks.updatePetProfile).toHaveBeenCalledWith({ ...pet, leaves: 350, levelUp: false });
    expect(invalidateQueries).not.toHaveBeenCalled();
  });
});
