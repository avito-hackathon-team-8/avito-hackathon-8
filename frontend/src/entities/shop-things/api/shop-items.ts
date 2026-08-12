import { apiRequest, getAuthHeaders } from '@/shared/api';
import { API_URL } from '@/shared/config';

import { API_ROUTE_SHOP } from './api-routes';

export type TShopItemCategory = 'BOWL' | 'BED';

export type TShopItemStatus = 'ACTIVE' | 'AVAILABLE' | 'LOCKED';

export type TShopItem = {
  id: string;
  category: TShopItemCategory;
  status: TShopItemStatus;
  title: string;
  description: string;
  imageUrl: string;
  requiredLevel: number;
  priceLeaves: number;
  durationDays: number;
};

type TResponseGetShopItems = {
  items: TShopItem[];
};

export type TPurchaseShopItemBody = {
  confirmReplacement: boolean;
};

export const getShopItems = async (): Promise<TShopItem[]> => {
  const data = await apiRequest<TResponseGetShopItems>(
    fetch(`${API_URL}${API_ROUTE_SHOP.items}`, {
      headers: getAuthHeaders(),
    }),
  );

  return data.items;
};

export const purchaseShopItem = async (
  itemId: TShopItem['id'],
  body: TPurchaseShopItemBody,
): Promise<void> => {
  return apiRequest<void>(
    fetch(`${API_URL}${API_ROUTE_SHOP.purchase(itemId)}`, {
      method: 'POST',
      headers: {
        ...getAuthHeaders(),
        'Content-Type': 'application/json',
      },
      body: JSON.stringify(body),
    }),
  );
};
