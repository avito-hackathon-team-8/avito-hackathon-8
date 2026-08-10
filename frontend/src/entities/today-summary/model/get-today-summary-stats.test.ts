import { describe, expect, it } from 'vitest';

import type { TTodaySummary } from '../api/get-today-summary';

import { getTodaySummaryStats } from './get-today-summary-stats ';

const createSummary = (overrides: Partial<TTodaySummary> = {}): TTodaySummary => ({
  leavesEarnedToday: 75,
  date: '2026-08-10',
  rewards: [],
  levelUp: null,
  tasks: [],
  visitedToday: true,
  updatedAt: '2026-08-10T12:00:00Z',
  ...overrides,
});

describe('getTodaySummaryStats', () => {
  it('объединяет события и сортирует их от новых к старым', () => {
    const summary = createSummary({
      tasks: [
        {
          taskId: 'task-1',
          type: 'VIEW_LISTINGS',
          description: 'Посмотреть объявления',
          rewardLeaves: 25,
          completedAt: '2026-08-10T09:00:00Z',
        },
      ],
      rewards: [
        {
          rewardId: 'reward-1',
          type: 'FREE_DELIVERY',
          title: 'Бесплатная доставка',
          expiresAt: '2026-08-20T10:00:00Z',
          receivedAt: '2026-08-10T11:00:00Z',
        },
      ],
      levelUp: {
        fromLevel: 2,
        toLevel: 3,
        occurredAt: '2026-08-10T10:00:00Z',
      },
    });

    const result = getTodaySummaryStats(summary);

    expect(result).toMatchObject({ leavesCount: 75, activitiesCount: 3 });
    expect(result.events.map(({ eventType }) => eventType)).toEqual(['REWARD', 'LEVEL_UP', 'TASK']);
  });

  it('возвращает пустой список событий для пустой сводки', () => {
    expect(getTodaySummaryStats(createSummary())).toEqual({
      leavesCount: 75,
      activitiesCount: 0,
      events: [],
    });
  });
});
