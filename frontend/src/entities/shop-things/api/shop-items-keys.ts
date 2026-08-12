import { mainQueryKey } from '@/shared/config';

export const shopItemsQueryKeys = {
  all: [mainQueryKey.all, 'shop-items'] as const,

  list: () => [...shopItemsQueryKeys.all, 'list'] as const,
};
