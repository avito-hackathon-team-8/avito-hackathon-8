import { render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { beforeEach, describe, expect, it, vi } from 'vitest';

import type { BottomPanelProps } from '@/shared/ui/bottom-panel/bottom-panel';

import type { TTodaySummaryStats } from '../../model/get-today-summary-stats ';

import { TodaySummaryCard } from './today-summary-card';

const mocks = vi.hoisted(() => ({
  useTodaySummary: vi.fn(),
  refetch: vi.fn(),
}));

vi.mock('../../model/use-today-summary', () => ({
  useTodaySummary: mocks.useTodaySummary,
}));

type BottomPanelStubProps = Pick<BottomPanelProps, 'children' | 'onClick' | 'renderTrigger'>;

vi.mock('@/shared/ui/bottom-panel', () => ({
  BottomPanel: ({ children, onClick, renderTrigger }: BottomPanelStubProps) => (
    <div>
      {renderTrigger(() => onClick?.())}
      {children}
    </div>
  ),
}));

const emptySummary: TTodaySummaryStats = {
  leavesCount: 0,
  activitiesCount: 0,
  events: [],
};

describe('TodaySummaryCard', () => {
  beforeEach(() => {
    mocks.useTodaySummary.mockReset();
    mocks.refetch.mockReset();
  });

  it('повторяет запрос по клику без данных', async () => {
    const user = userEvent.setup();
    mocks.useTodaySummary.mockReturnValue({ data: undefined, refetch: mocks.refetch });
    render(<TodaySummaryCard />);

    await user.click(screen.getByRole('article'));
    expect(mocks.refetch).toHaveBeenCalledOnce();
  });

  it('показывает empty-state для пустой сводки', () => {
    mocks.useTodaySummary.mockReturnValue({ data: emptySummary, refetch: mocks.refetch });
    render(<TodaySummaryCard />);

    expect(screen.getByText('Активностей: 0 · +0 листьев')).toBeInTheDocument();
    expect(
      screen.getByRole('heading', { name: 'Сегодня пока нет активности' }),
    ).toBeInTheDocument();
  });

  it('показывает число активностей и содержимое непустой сводки', () => {
    const summary: TTodaySummaryStats = {
      leavesCount: 25,
      activitiesCount: 1,
      events: [
        {
          eventType: 'TASK',
          occurredAt: '2026-08-10T10:00:00Z',
          data: {
            taskId: 'task-1',
            type: 'VIEW_LISTINGS',
            description: 'Посмотреть объявления',
            rewardLeaves: 25,
            completedAt: '2026-08-10T10:00:00Z',
          },
        },
      ],
    };
    mocks.useTodaySummary.mockReturnValue({ data: summary, refetch: mocks.refetch });
    render(<TodaySummaryCard />);

    expect(screen.getByText('Активность: 1 · +25 листьев')).toBeInTheDocument();
    expect(screen.getByRole('heading', { name: 'Посмотреть объявления' })).toBeInTheDocument();
  });
});
