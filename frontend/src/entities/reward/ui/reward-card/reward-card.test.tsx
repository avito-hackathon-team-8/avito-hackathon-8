import { render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { beforeEach, describe, expect, it, vi } from 'vitest';

import type { BottomPanelProps } from '@/shared/ui/bottom-panel/bottom-panel';

import type { TReward } from '../../api/rewards';

import { RewardCard } from './reward-card';

const mocks = vi.hoisted(() => ({
  useRewards: vi.fn(),
  refetch: vi.fn(),
}));

vi.mock('../../model/use-reward', () => ({
  useRewards: mocks.useRewards,
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

const createReward = (id: string): TReward => ({
  id,
  title: `Награда ${id}`,
  category: 'FREE_DELIVERY',
  categoryName: 'Доставка',
  source: 'CHEST',
  active: true,
  status: 'ACTIVE',
  expiresAt: '2026-08-20T10:00:00Z',
  awardedAt: '2026-08-10T10:00:00Z',
  redeemedAt: null,
});

describe('RewardCard', () => {
  beforeEach(() => {
    mocks.useRewards.mockReset();
    mocks.refetch.mockReset();
  });

  it('показывает ноль и повторяет запрос без данных', async () => {
    const user = userEvent.setup();
    mocks.useRewards.mockReturnValue({
      data: undefined,
      isPending: false,
      refetch: mocks.refetch,
    });
    render(<RewardCard />);

    expect(screen.getByText('0 бонусов доступно')).toBeInTheDocument();
    await user.click(screen.getByRole('article'));
    expect(mocks.refetch).toHaveBeenCalledOnce();
  });

  it('не повторяет запрос во время загрузки', async () => {
    const user = userEvent.setup();
    mocks.useRewards.mockReturnValue({
      data: undefined,
      isPending: true,
      refetch: mocks.refetch,
    });
    render(<RewardCard />);

    await user.click(screen.getByRole('article'));
    expect(mocks.refetch).not.toHaveBeenCalled();
  });

  it('объединяет группы и показывает корректное количество наград', () => {
    mocks.useRewards.mockReturnValue({
      data: {
        groups: [
          { category: 'FREE_DELIVERY', categoryName: 'Доставка', items: [createReward('1')] },
          { category: 'FREE_PROMOTION', categoryName: 'Продвижение', items: [createReward('2')] },
        ],
      },
      isPending: false,
      refetch: mocks.refetch,
    });
    render(<RewardCard />);

    expect(screen.getByText('2 бонуса доступны')).toBeInTheDocument();
    expect(screen.getAllByRole('link', { name: 'Применить' })).toHaveLength(2);
  });
});
