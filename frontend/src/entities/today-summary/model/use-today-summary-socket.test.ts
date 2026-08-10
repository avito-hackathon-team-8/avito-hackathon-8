import { act, renderHook } from '@testing-library/react';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';

import { sessionStorageKeysMap } from '@/shared/lib';
import { MockWebSocket } from '@/test/mock-web-socket';
import { createQueryWrapper, createTestQueryClient } from '@/test/render-with-providers';

import type { TTodaySummary } from '../api/get-today-summary';
import { todaySummaryQueryKeys } from '../api/today-summary-query-keys';

import { useTodaySummarySocket } from './use-today-summary-socket';

const summary: TTodaySummary = {
  leavesEarnedToday: 50,
  date: '2026-08-10',
  rewards: [],
  levelUp: null,
  tasks: [],
  visitedToday: true,
  updatedAt: '2026-08-10T12:00:00Z',
};

describe('useTodaySummarySocket', () => {
  beforeEach(() => {
    vi.useFakeTimers();
    vi.stubGlobal('WebSocket', MockWebSocket);
    MockWebSocket.reset();
    vi.spyOn(console, 'error').mockImplementation(() => undefined);
    vi.spyOn(console, 'warn').mockImplementation(() => undefined);
  });

  afterEach(() => {
    vi.useRealTimers();
    vi.unstubAllGlobals();
  });

  it('не подключается без токена', async () => {
    const queryClient = createTestQueryClient();
    renderHook(() => useTodaySummarySocket(), {
      wrapper: createQueryWrapper(queryClient),
    });

    await act(async () => vi.runOnlyPendingTimersAsync());
    expect(MockWebSocket.instances).toHaveLength(0);
  });

  it('подключается с токеном и обновляет кэш сводки', async () => {
    sessionStorage.setItem(sessionStorageKeysMap.authToken, 'token-1');
    const queryClient = createTestQueryClient();
    renderHook(() => useTodaySummarySocket(), {
      wrapper: createQueryWrapper(queryClient),
    });

    await act(async () => vi.advanceTimersByTimeAsync(0));
    const socket = MockWebSocket.instances[0];
    const url = new URL(socket.url);

    expect(url.protocol).toBe('ws:');
    expect(url.pathname).toBe('/api/v1/daily-report/ws');
    expect(url.searchParams.get('token')).toBe('token-1');

    act(() => {
      socket.emitMessage({ event: 'DAILY_REPORT_UPDATED', data: summary });
    });
    expect(queryClient.getQueryData(todaySummaryQueryKeys.current())).toEqual(summary);
  });

  it('игнорирует неизвестные события и сообщает о повреждённом JSON', async () => {
    sessionStorage.setItem(sessionStorageKeysMap.authToken, 'token-1');
    const queryClient = createTestQueryClient();
    queryClient.setQueryData(todaySummaryQueryKeys.current(), summary);
    renderHook(() => useTodaySummarySocket(), {
      wrapper: createQueryWrapper(queryClient),
    });
    await act(async () => vi.advanceTimersByTimeAsync(0));
    const socket = MockWebSocket.instances[0];

    act(() => socket.emitMessage({ event: 'UNKNOWN', data: { leavesEarnedToday: 999 } }));
    expect(queryClient.getQueryData(todaySummaryQueryKeys.current())).toEqual(summary);

    act(() => socket.emitMessage('{broken-json'));
    expect(console.error).toHaveBeenCalledWith(
      'Ошибка разбора ежедневной сводки из WebSocket:',
      expect.any(SyntaxError),
    );
  });

  it('переподключается после обрыва и закрывает сокет при unmount', async () => {
    sessionStorage.setItem(sessionStorageKeysMap.authToken, 'token-1');
    const queryClient = createTestQueryClient();
    const { unmount } = renderHook(() => useTodaySummarySocket(), {
      wrapper: createQueryWrapper(queryClient),
    });
    await act(async () => vi.advanceTimersByTimeAsync(0));
    const firstSocket = MockWebSocket.instances[0];

    act(() => {
      firstSocket.emitOpen();
      firstSocket.emitClose(1006);
    });
    await act(async () => vi.advanceTimersByTimeAsync(999));
    expect(MockWebSocket.instances).toHaveLength(1);

    await act(async () => vi.advanceTimersByTimeAsync(1));
    expect(MockWebSocket.instances).toHaveLength(2);
    const secondSocket = MockWebSocket.instances[1];

    unmount();
    expect(secondSocket.close).toHaveBeenCalledWith(1000, 'Component unmounted');

    await act(async () => vi.runOnlyPendingTimersAsync());
    expect(MockWebSocket.instances).toHaveLength(2);
  });

  it('не переподключается, если первоначальное соединение не открылось', async () => {
    sessionStorage.setItem(sessionStorageKeysMap.authToken, 'token-1');
    const queryClient = createTestQueryClient();
    renderHook(() => useTodaySummarySocket(), {
      wrapper: createQueryWrapper(queryClient),
    });
    await act(async () => vi.advanceTimersByTimeAsync(0));

    act(() => MockWebSocket.instances[0].emitClose(1006));
    await act(async () => vi.advanceTimersByTimeAsync(10_000));

    expect(MockWebSocket.instances).toHaveLength(1);
    expect(console.error).toHaveBeenCalledWith(
      'WebSocket не смог установить первоначальное соединение',
    );
  });
});
