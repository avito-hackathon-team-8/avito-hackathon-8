import { mainQueryKey } from '@/shared/config/api';

export const tasksQueryKeys = {
  all: [mainQueryKey.all, 'tasks'] as const,

  list: () => [...tasksQueryKeys.all, 'list'] as const,
};
