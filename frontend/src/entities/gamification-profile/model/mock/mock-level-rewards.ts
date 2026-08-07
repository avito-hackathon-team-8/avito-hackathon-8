import type { TLevelRewardItem } from '../../api/get-levels-reward';

export const mockLevelRewards: TLevelRewardItem[] = [
  {
    level: 1,
    status: 'CLAIMED',
    reward: {
      id: '10101010-1010-4010-8010-101010101010',
      type: 'AVITO_BONUS',
      description: '50 бонусов Авито',
    },
    expiresAt: null,
  },
  {
    level: 2,
    status: 'CLAIMED',
    reward: {
      id: '20202020-2020-4020-8020-202020202020',
      type: 'FREE_DELIVERY',
      description: 'Бесплатная доставка для одного заказа',
    },
    expiresAt: null,
  },
  {
    level: 3,
    status: 'FROZEN',
    reward: {
      id: '30303030-3030-4030-8030-303030303030',
      type: 'PROMOTION_DISCOUNT',
      description: 'Скидка 10% на продвижение объявления',
    },
    expiresAt: '2026-08-03T12:00:00Z',
  },
  {
    level: 4,
    status: 'UNOPENED',
    reward: {
      id: '40404040-4040-4040-8040-404040404040',
      type: 'FREE_PROMOTION',
      description: 'Бесплатное продвижение объявления на 3 дня',
    },
    expiresAt: '2026-08-09T12:00:00Z',
  },
  {
    level: 5,
    status: 'LOCKED',
    reward: {
      id: '50505050-5050-4050-8050-505050505050',
      type: 'DELIVERY_DISCOUNT',
      description: 'Скидка 15% на Авито Доставку',
    },
    expiresAt: null,
  },
  {
    level: 6,
    status: 'LOCKED',
    reward: {
      id: '60606060-6060-4060-8060-606060606060',
      type: 'AVITO_BONUS',
      description: '100 бонусов Авито',
    },
    expiresAt: null,
  },
  {
    level: 7,
    status: 'LOCKED',
    reward: {
      id: '70707070-7070-4070-8070-707070707070',
      type: 'FREE_DELIVERY',
      description: 'Бесплатная доставка для двух заказов',
    },
    expiresAt: null,
  },
  {
    level: 8,
    status: 'LOCKED',
    reward: {
      id: '80808080-8080-4080-8080-808080808080',
      type: 'PROMOTION_DISCOUNT',
      description: 'Скидка 20% на продвижение объявления',
    },
    expiresAt: null,
  },
  {
    level: 9,
    status: 'LOCKED',
    reward: {
      id: '90909090-9090-4090-8090-909090909090',
      type: 'FREE_PROMOTION',
      description: 'Бесплатное продвижение объявления на 7 дней',
    },
    expiresAt: null,
  },
  {
    level: 10,
    status: 'LOCKED',
    reward: {
      id: 'a0a0a0a0-a0a0-40a0-80a0-a0a0a0a0a0a0',
      type: 'DELIVERY_DISCOUNT',
      description: 'Скидка 25% на Авито Доставку',
    },
    expiresAt: null,
  },
];
