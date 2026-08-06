import type { TRewardGroup } from '../../api/get-rewards';

export const mockRewards: TRewardGroup[] = [
  {
    category: 'AVITO_BONUS',
    categoryName: 'Бонусы Авито',
    items: [
      {
        id: '49a70fa1-4a2a-4cc7-b09d-444444444444',
        title: '100 бонусов',
        category: 'FREE_DELIVERY',
        categoryName: 'Бонусы Авито',
        source: 'LEVEL',
        active: true,
        status: 'ACTIVE',
        expiresAt: '2026-09-04T12:00:00Z',
        awardedAt: '2026-08-05T12:00:00Z',
        redeemedAt: null,
      },
      {
        id: '39bc5134-b86d-40cb-b162-666666666666',
        title: '200 бонусов',
        category: 'AVITO_BONUS',
        categoryName: 'Бонусы Авито',
        source: 'CHEST',
        active: false,
        status: 'EXPIRED',
        expiresAt: '2026-08-01T12:00:00Z',
        awardedAt: '2026-07-05T12:00:00Z',
        redeemedAt: null,
      },
    ],
  },
  {
    category: 'FREE_DELIVERY',
    categoryName: 'Бесплатная доставка',
    items: [
      {
        id: '2c9f7f9f-a958-4e77-b17c-555555555555',
        title: 'Бесплатная доставка (1 заказ)',
        category: 'FREE_DELIVERY',
        categoryName: 'Бесплатная доставка',
        source: 'CHEST',
        active: false,
        status: 'REDEEMED',
        expiresAt: '2026-08-20T12:00:00Z',
        awardedAt: '2026-08-05T12:00:00Z',
        redeemedAt: '2026-08-05T13:30:00Z',
      },
      {
        id: '8a41bb48-dcf7-4db9-b149-777777777777',
        title: 'Бесплатная доставка (2 заказа)',
        category: 'FREE_DELIVERY',
        categoryName: 'Бесплатная доставка',
        source: 'LEVEL',
        active: true,
        status: 'ACTIVE',
        expiresAt: '2026-09-05T12:00:00Z',
        awardedAt: '2026-08-06T12:00:00Z',
        redeemedAt: null,
      },
    ],
  },
];
