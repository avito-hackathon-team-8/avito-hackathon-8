import { mainQueryKey } from '@/shared/config';

export const activityDayKeys = {
  all: [mainQueryKey.all, 'activity-days'] as const,

  week: () => [...activityDayKeys.all, 'current'] as const,
};
