import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { toast } from 'sonner';

import { rewardsQueryKeys } from '@/entities/reward';

import { getShopItems, purchaseShopItem, type TShopItem } from '../api/shop-items';
import { shopItemsQueryKeys } from '../api/shop-items-keys';

export type TPurchaseShopItemVariables = {
  itemId: TShopItem['id'];
  confirmReplacement?: boolean;
};

const PURCHASE_ERROR_MESSAGE = 'Не удалось купить товар';

const getPurchaseErrorMessage = (error: unknown): string => {
  if (!(error instanceof Error) || !error.message) {
    return PURCHASE_ERROR_MESSAGE;
  }

  try {
    const parsedError = JSON.parse(error.message) as { message?: unknown };

    if (typeof parsedError.message === 'string' && parsedError.message) {
      return parsedError.message;
    }
  } catch {
    return error.message;
  }

  return error.message;
};

export const useShopItems = () => {
  const queryClient = useQueryClient();

  const queryKey = shopItemsQueryKeys.list();

  const shopItemsQuery = useQuery({
    queryKey,
    queryFn: getShopItems,
    retry: 1,
  });

  const purchaseMutation = useMutation({
    mutationFn: ({ itemId, confirmReplacement = false }: TPurchaseShopItemVariables) =>
      purchaseShopItem(itemId, { confirmReplacement }),

    onError: (error) => {
      toast.error(getPurchaseErrorMessage(error));
    },

    onSuccess: () => {
      return Promise.all([
        queryClient.invalidateQueries({ queryKey }),
        queryClient.invalidateQueries({ queryKey: rewardsQueryKeys.list() }),
      ]);
    },
  });

  return {
    ...shopItemsQuery,

    purchaseItem: purchaseMutation.mutate,
    isPurchasePending: purchaseMutation.isPending,
  };
};
