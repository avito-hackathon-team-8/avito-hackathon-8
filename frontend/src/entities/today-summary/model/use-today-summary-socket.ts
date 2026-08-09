import { useEffect } from 'react';

import { useQueryClient } from '@tanstack/react-query';

import { API_URL } from '@/shared/config';
import { getSessionStorageValue, sessionStorageKeysMap } from '@/shared/lib';

import { API_ROUTE_TODAY_SUMMARY } from '../api/api-routes';
import type { TTodaySummary } from '../api/get-today-summary';
import { todaySummaryQueryKeys } from '../api/today-summary-query-keys';

type TTodaySummarySocketEvent = {
  event: 'DAILY_REPORT_UPDATED';
  data: TTodaySummary;
};

const INITIAL_RECONNECT_DELAY = 1_000;
const MAX_RECONNECT_DELAY = 10_000;

export const useTodaySummarySocket = () => {
  const queryClient = useQueryClient();

  useEffect(() => {
    const token = getSessionStorageValue(sessionStorageKeysMap.authToken);

    if (!token) {
      return;
    }

    let socket: WebSocket | null = null;

    let reconnectTimer: ReturnType<typeof setTimeout> | null = null;
    let connectTimer: ReturnType<typeof setTimeout> | null = null;

    let reconnectDelay = INITIAL_RECONNECT_DELAY;

    let disposed = false;
    let wasConnected = false;

    const connect = () => {
      if (disposed) {
        return;
      }

      const socketUrl = new URL(`${API_URL}${API_ROUTE_TODAY_SUMMARY.reportWs}`);

      socketUrl.protocol = socketUrl.protocol === 'https:' ? 'wss:' : 'ws:';

      socketUrl.searchParams.set('token', token);

      socket = new WebSocket(socketUrl);

      socket.onopen = () => {
        wasConnected = true;
        reconnectDelay = INITIAL_RECONNECT_DELAY;
      };

      socket.onmessage = (event) => {
        try {
          const payload = JSON.parse(event.data) as TTodaySummarySocketEvent;

          if (payload.event !== 'DAILY_REPORT_UPDATED') {
            return;
          }

          queryClient.setQueryData(todaySummaryQueryKeys.current(), payload.data);
        } catch (error) {
          console.error('Ошибка разбора ежедневной сводки из WebSocket:', error);
        }
      };

      socket.onerror = () => {
        console.warn('Ошибка WebSocket ежедневной сводки');
      };

      socket.onclose = (event) => {
        if (disposed) {
          return;
        }

        if (!wasConnected) {
          console.error('WebSocket не смог установить первоначальное соединение');

          return;
        }

        if (event.code === 1000) {
          return;
        }

        reconnectTimer = setTimeout(() => {
          connect();
        }, reconnectDelay);

        reconnectDelay = Math.min(reconnectDelay * 2, MAX_RECONNECT_DELAY);
      };
    };

    connectTimer = setTimeout(connect, 0);

    return () => {
      disposed = true;

      if (connectTimer) {
        clearTimeout(connectTimer);
      }

      if (reconnectTimer) {
        clearTimeout(reconnectTimer);
      }

      socket?.close(1000, 'Component unmounted');
    };
  }, [queryClient]);
};
