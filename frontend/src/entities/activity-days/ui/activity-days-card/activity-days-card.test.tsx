import { render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { beforeEach, describe, expect, it, vi } from 'vitest';

import type { BottomPanelProps } from '@/shared/ui/bottom-panel/bottom-panel';

import type { TResponseActivityDay } from '../../api/activity-day';

import { ActivityDaysCard } from './activity-days-card';

const mocks = vi.hoisted(() => ({
  useActivityDays: vi.fn(),
  refetch: vi.fn(),
  receiveReward: vi.fn(),
}));

vi.mock('../../model/use-activity-days', () => ({
  useActivityDays: mocks.useActivityDays,
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

const week: TResponseActivityDay = {
  claimedDaysCount: 2,
  claims: [
    {
      weekday: 2,
      date: '2026-08-11',
      status: 'AVAILABLE',
      rewardLeaves: 20,
      claimId: 'claim-2',
    },
  ],
};

describe('ActivityDaysCard', () => {
  beforeEach(() => {
    mocks.useActivityDays.mockReset();
    mocks.refetch.mockReset();
    mocks.receiveReward.mockReset();
  });

  it('повторяет запрос по клику, если данных нет', async () => {
    const user = userEvent.setup();
    mocks.useActivityDays.mockReturnValue({
      data: undefined,
      isPending: false,
      refetch: mocks.refetch,
      receiveReward: mocks.receiveReward,
    });
    render(<ActivityDaysCard />);

    expect(screen.getByText('Не удалось получить данные')).toBeInTheDocument();
    await user.click(screen.getByRole('article'));
    expect(mocks.refetch).toHaveBeenCalledOnce();
  });

  it('не повторяет запрос во время загрузки', async () => {
    const user = userEvent.setup();
    mocks.useActivityDays.mockReturnValue({
      data: undefined,
      isPending: true,
      refetch: mocks.refetch,
      receiveReward: mocks.receiveReward,
    });
    render(<ActivityDaysCard />);

    await user.click(screen.getByRole('article'));
    expect(mocks.refetch).not.toHaveBeenCalled();
  });

  it('показывает серию и передаёт получение награды', async () => {
    const user = userEvent.setup();
    mocks.useActivityDays.mockReturnValue({
      data: week,
      isPending: false,
      refetch: mocks.refetch,
      receiveReward: mocks.receiveReward,
    });
    render(<ActivityDaysCard />);

    expect(screen.getByText('2 дня на этой неделе')).toBeInTheDocument();
    await user.click(screen.getByRole('button', { name: 'вт 20' }));
    expect(mocks.receiveReward).toHaveBeenCalledOnce();
    expect(mocks.refetch).not.toHaveBeenCalled();
  });
});
