import { render, screen } from '@testing-library/react';
import { describe, expect, it } from 'vitest';

import type { TTodaySummaryStats } from '../../model/get-today-summary-stats ';

import { TodaySummaryPanelContent } from './today-summary-panel-content';

const events: TTodaySummaryStats = {
  leavesCount: 75,
  activitiesCount: 3,
  events: [
    {
      eventType: 'TASK',
      occurredAt: '2026-08-10T09:00:00Z',
      data: {
        taskId: 'task-1',
        type: 'VIEW_LISTINGS',
        description: 'Посмотреть объявления',
        rewardLeaves: 25,
        completedAt: '2026-08-10T09:00:00Z',
      },
    },
    {
      eventType: 'REWARD',
      occurredAt: '2026-08-10T10:00:00Z',
      data: {
        rewardId: 'reward-1',
        type: 'FREE_DELIVERY',
        title: 'Бесплатная доставка',
        expiresAt: '2026-08-20T10:00:00Z',
        receivedAt: '2026-08-10T10:00:00Z',
      },
    },
    {
      eventType: 'LEVEL_UP',
      occurredAt: '2026-08-10T11:00:00Z',
      data: {
        fromLevel: 2,
        toLevel: 3,
        occurredAt: '2026-08-10T11:00:00Z',
      },
    },
  ],
};

describe('TodaySummaryPanelContent', () => {
  it('показывает количество листьев и все типы событий', () => {
    render(<TodaySummaryPanelContent events={events} />);

    expect(screen.getByText('+ 75')).toBeInTheDocument();
    expect(screen.getByRole('heading', { name: 'Посмотреть объявления' })).toBeInTheDocument();
    expect(screen.getByRole('heading', { name: 'Бесплатная доставка' })).toBeInTheDocument();
    expect(screen.getByRole('heading', { name: 'Уровень повышен с 2 до 3' })).toBeInTheDocument();
    expect(screen.getByText('2026-08-20')).toBeInTheDocument();
  });

  it('корректно отображает пустой список событий', () => {
    render(
      <TodaySummaryPanelContent events={{ leavesCount: 0, activitiesCount: 0, events: [] }} />,
    );

    expect(screen.getByText('+ 0')).toBeInTheDocument();
    expect(screen.queryAllByRole('listitem')).toHaveLength(0);
  });
});
