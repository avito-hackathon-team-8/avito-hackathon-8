import { useEffect } from 'react';

import { useQueryClient } from '@tanstack/react-query';

import { dailyTasksQueryKeys } from '@/entities/daily-task';
import { API_URL } from '@/shared/config';
import { getSessionStorageValue, sessionStorageKeysMap } from '@/shared/lib';

import { API_ROUTE_PROFILE } from '../api/api-routes';
import { gamificationProfileKeys } from '../api/gamification-profile-keys';
import type { TPet } from '../api/pet';

import { usePetProfile } from './use-pet-profile';

type PetProgressSocketEvent = {
  event: 'PET_PROGRESS_UPDATED';
  data: TPet;
};

const INITIAL_RECONNECT_DELAY = 1_000;
const MAX_RECONNECT_DELAY = 10_000;

type TUsePetProfileSocketProps = {
  enabled: boolean;
};

export const usePetProfileSocket = ({ enabled }: TUsePetProfileSocketProps) => {
  const queryClient = useQueryClient();
  const { updatePetProfile } = usePetProfile();

  useEffect(() => {
    if (!enabled) return;

    const token = getSessionStorageValue(sessionStorageKeysMap.authToken);

    if (!token) {
      return;
    }

    let socket: WebSocket | null = null;

    let reconnectTimer: ReturnType<typeof setTimeout> | null = null;
    let connectTimer: ReturnType<typeof setTimeout> | null = null;

    let reconnectDelay = INITIAL_RECONNECT_DELAY;

    let disposed = false;

    const connect = () => {
      if (disposed) {
        return;
      }

      const socketUrl = new URL(`${API_URL}${API_ROUTE_PROFILE.petProfileWs}`);

      socketUrl.protocol = socketUrl.protocol === 'https:' ? 'wss:' : 'ws:';

      socketUrl.searchParams.set('token', token);

      socket = new WebSocket(socketUrl);

      socket.onopen = () => {
        reconnectDelay = INITIAL_RECONNECT_DELAY;
      };

      socket.onmessage = (event) => {
        try {
          const payload = JSON.parse(event.data) as PetProgressSocketEvent;

          if (payload.event !== 'PET_PROGRESS_UPDATED') {
            return;
          }

          const previousPet = queryClient.getQueryData<TPet>(gamificationProfileKeys.pet());

          const isLevelUpdated =
            payload.data.levelUp ||
            (previousPet !== undefined && payload.data.level > previousPet.level);

          updatePetProfile(payload.data);

          if (isLevelUpdated) {
            void Promise.all([
              queryClient.invalidateQueries({
                queryKey: gamificationProfileKeys.levels(),
              }),
              queryClient.invalidateQueries({
                queryKey: dailyTasksQueryKeys.list(),
              }),
            ]);
          }
        } catch (error) {
          console.error('Ошибка разбора события питомца из WebSocket:', error);
        }
      };

      socket.onerror = () => {
        console.warn('Ошибка WebSocket питомца');
      };

      socket.onclose = (event) => {
        console.log('Pet WebSocket закрыт', {
          code: event.code,
          reason: event.reason,
          wasClean: event.wasClean,
        });

        if (disposed) {
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
  }, [enabled, queryClient, updatePetProfile]);
};
