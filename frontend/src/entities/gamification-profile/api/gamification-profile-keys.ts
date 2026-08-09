import { mainQueryKey } from '@/shared/config';

export const gamificationProfileKeys = {
  all: [...mainQueryKey.all, 'gamification-profile'] as const,
  pet: () => [...gamificationProfileKeys.all, 'pet'] as const,
  levels: () => [...gamificationProfileKeys.all, 'levels'] as const,
};
