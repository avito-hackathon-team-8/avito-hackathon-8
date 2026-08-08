import { mainQueryKey } from '@/shared/config/api';

export const todaySummaryQueryKeys = {
  all: [mainQueryKey.all, 'today-summary'] as const,
  current: () => [...todaySummaryQueryKeys.all, 'current'] as const,
};
