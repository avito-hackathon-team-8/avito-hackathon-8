import { act, renderHook } from '@testing-library/react';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';

import { dailyTasksQueryKeys } from '@/entities/daily-task';
import { shopItemsQueryKeys } from '@/entities/shop-things';
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
  bowlImageUrl: null,
  bedImageUrl: null,
  happiness: 50,
  happinessMultiplier: 1,
  calculatedAt: '2026-08-12T12:52:25.179950567Z',
  decaysToZeroAt: '2026-08-15T12:52:15.223227999Z',
  feedNextAvailableAt: null,
  strokeNextAvailableAt: null,
};

const petProgress = {
  chestPrice: pet.chestPrice,
  leaves: pet.leaves,
  level: pet.level,
  levelUp: false,
  name: pet.name,
  nextLevelTargetLeaves: pet.nextLevelTargetLeaves,
  bowlImageUrl: '/api/v1/shop-images/bowl-fashionable.webp',
  bedImageUrl: '/api/v1/shop-images/bed-car.webp',
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
        data: petProgress,
      });
    });

    expect(mocks.updatePetProfile).toHaveBeenCalledWith({
      ...pet,
      ...petProgress,
    });
    expect(invalidateQueries).toHaveBeenCalledWith({
      queryKey: gamificationProfileKeys.levels(),
    });
    expect(invalidateQueries).toHaveBeenCalledWith({ queryKey: dailyTasksQueryKeys.list() });
    expect(invalidateQueries).toHaveBeenCalledWith({ queryKey: shopItemsQueryKeys.list() });
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
        data: { ...petProgress, leaves: 350 },
      });
    });

    expect(mocks.updatePetProfile).toHaveBeenCalledWith({
      ...pet,
      ...petProgress,
      leaves: 350,
    });
    expect(invalidateQueries).not.toHaveBeenCalled();
  });

  it('ревалидирует профиль при обновлении состояния питомца', async () => {
    sessionStorage.setItem(sessionStorageKeysMap.authToken, 'token-1');
    const queryClient = createTestQueryClient();
    const invalidateQueries = vi.spyOn(queryClient, 'invalidateQueries').mockResolvedValue();
    renderHook(() => usePetProfileSocket({ enabled: true }), {
      wrapper: createQueryWrapper(queryClient),
    });
    await act(async () => vi.advanceTimersByTimeAsync(0));

    act(() => {
      MockWebSocket.instances[0].emitMessage({
        event: 'PET_STATE_UPDATED',
        data: {
          calculatedAt: '2026-08-12T13:12:40.413394178Z',
          decaysToZeroAt: '2026-08-15T12:52:15.223228Z',
          feedNextAvailableAt: null,
          happiness: 99.53,
          happinessMultiplier: 1.495,
          strokeNextAvailableAt: null,
        },
      });
    });

    expect(invalidateQueries).toHaveBeenCalledWith({
      queryKey: gamificationProfileKeys.pet(),
    });
    expect(mocks.updatePetProfile).not.toHaveBeenCalled();
  });
});
