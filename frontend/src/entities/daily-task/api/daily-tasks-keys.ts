import { mainQueryKey } from '@/shared/config';

export const dailyTasksQueryKeys = {
  all: [mainQueryKey.all, 'tasks'] as const,

  list: () => [...dailyTasksQueryKeys.all, 'list'] as const,
};
