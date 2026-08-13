import { act, renderHook, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';

import { rewardsQueryKeys } from '@/entities/reward';
import { createQueryWrapper, createTestQueryClient } from '@/test/render-with-providers';

import type { TShopItem } from '../api/shop-items';
import { shopItemsQueryKeys } from '../api/shop-items-keys';

import { useShopItems } from './use-shop-items';

const mocks = vi.hoisted(() => ({
  getShopItems: vi.fn(),
  purchaseShopItem: vi.fn(),
  toastError: vi.fn(),
}));

vi.mock('../api/shop-items', () => ({
  getShopItems: mocks.getShopItems,
  purchaseShopItem: mocks.purchaseShopItem,
}));

vi.mock('sonner', () => ({
  toast: { error: mocks.toastError },
}));

const shopItem: TShopItem = {
  id: 'fashionable-bowl',
  category: 'BOWL',
  status: 'AVAILABLE',
  title: 'Модная миска',
  description: '5% кэшбека листьями за покупки.',
  imageUrl: '/api/v1/shop-images/bowl-fashionable.webp',
  requiredLevel: 5,
  priceLeaves: 100,
  durationDays: 3,
};

describe('useShopItems', () => {
  beforeEach(() => {
    mocks.getShopItems.mockReset().mockResolvedValue([shopItem]);
    mocks.purchaseShopItem.mockReset();
    mocks.toastError.mockReset();
  });

  it('получает товары и после покупки обновляет магазин и награды', async () => {
    mocks.purchaseShopItem.mockResolvedValue(undefined);
    const queryClient = createTestQueryClient();
    const invalidateQueries = vi.spyOn(queryClient, 'invalidateQueries');
    const { result } = renderHook(() => useShopItems(), {
      wrapper: createQueryWrapper(queryClient),
    });

    await waitFor(() => expect(result.current.data).toEqual([shopItem]));

    act(() => result.current.purchaseItem({ itemId: shopItem.id }));

    await waitFor(() => {
      expect(mocks.purchaseShopItem).toHaveBeenCalledWith(shopItem.id, {
        confirmReplacement: false,
      });
    });
    await waitFor(() => {
      expect(invalidateQueries).toHaveBeenCalledWith({ queryKey: shopItemsQueryKeys.list() });
      expect(invalidateQueries).toHaveBeenCalledWith({ queryKey: rewardsQueryKeys.list() });
    });
  });

  it('передаёт подтверждение замены активного товара', async () => {
    mocks.purchaseShopItem.mockResolvedValue(undefined);
    const { result } = renderHook(() => useShopItems(), {
      wrapper: createQueryWrapper(),
    });

    act(() =>
      result.current.purchaseItem({
        itemId: shopItem.id,
        confirmReplacement: true,
      }),
    );

    await waitFor(() => {
      expect(mocks.purchaseShopItem).toHaveBeenCalledWith(shopItem.id, {
        confirmReplacement: true,
      });
    });
  });

  it('показывает сообщение backend при ошибке покупки', async () => {
    mocks.purchaseShopItem.mockRejectedValue(
      new Error(
        JSON.stringify({
          code: 'INSUFFICIENT_LEAVES',
          message: 'Недостаточно листьев для покупки предмета',
        }),
      ),
    );
    const { result } = renderHook(() => useShopItems(), {
      wrapper: createQueryWrapper(),
    });

    act(() => result.current.purchaseItem({ itemId: shopItem.id }));

    await waitFor(() => {
      expect(mocks.toastError).toHaveBeenCalledWith('Недостаточно листьев для покупки предмета');
    });
  });

  it('показывает текст обычной ошибки покупки', async () => {
    mocks.purchaseShopItem.mockRejectedValue(new Error('Сервис временно недоступен'));
    const { result } = renderHook(() => useShopItems(), {
      wrapper: createQueryWrapper(),
    });

    act(() => result.current.purchaseItem({ itemId: shopItem.id }));

    await waitFor(() => {
      expect(mocks.toastError).toHaveBeenCalledWith('Сервис временно недоступен');
    });
  });

  it('показывает fallback при неизвестной ошибке покупки', async () => {
    mocks.purchaseShopItem.mockRejectedValue(null);
    const { result } = renderHook(() => useShopItems(), {
      wrapper: createQueryWrapper(),
    });

    act(() => result.current.purchaseItem({ itemId: shopItem.id }));

    await waitFor(() => {
      expect(mocks.toastError).toHaveBeenCalledWith('Не удалось купить товар');
    });
  });
});
