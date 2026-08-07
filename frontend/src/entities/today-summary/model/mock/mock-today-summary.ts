import type { TTodaySummary } from '../../api/get-today-summary';

export const mockTodaySummary: TTodaySummary = {
  leavesEarnedToday: 150,
  date: '2026-08-05',
  rewards: [
    {
      rewardId: '96b41fc5-1a5f-4db4-9129-555555555555',
      type: 'PROMOTION_DISCOUNT',
      title: 'Скидка 20% на продвижение',
      expiresAt: '2026-09-01',
      receivedAt: '2026-08-05T10:00:00Z',
    },
  ],
  levelUp: {
    fromLevel: 3,
    toLevel: 4,
    occurredAt: '2026-08-05T10:00:00Z',
  },
  tasks: [
    {
      taskId: '8f123af2-24c5-4ff7-b114-111111111111',
      type: 'VIEW_LISTINGS',
      description: 'Посмотреть 5 объявлений',
      rewardLeaves: 45,
      rewardClaimed: true,
      completedAt: '2026-08-05T09:15:00Z',
    },
    {
      taskId: '6ab41fc5-1a5f-4db4-9129-222222222222',
      type: 'ADD_TO_FAVORITES',
      description: 'Добавить 3 объявления в избранное',
      rewardLeaves: 45,
      rewardClaimed: false,
      completedAt: '2026-08-05T11:20:00Z',
    },
  ],
  visitedToday: true,
  updatedAt: '2026-08-05T11:20:00Z',
};
