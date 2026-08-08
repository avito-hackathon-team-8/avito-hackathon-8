import { mainQueryKey } from '@/shared/config/api';

export const rewardsQueryKeys = {
  all: [mainQueryKey.all, 'rewards'] as const,

  list: () => [...rewardsQueryKeys.all, 'list'] as const,
};
