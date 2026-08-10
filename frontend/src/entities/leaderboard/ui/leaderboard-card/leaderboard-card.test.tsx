import { render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { beforeEach, describe, expect, it, vi } from 'vitest';

import type { User } from '@/entities/user';
import type { BottomPanelProps } from '@/shared/ui/bottom-panel/bottom-panel';

import { LeaderboardCard } from './leaderboard-card';

const mocks = vi.hoisted(() => ({
  useCurrentUser: vi.fn(),
  useLeaderboard: vi.fn(),
  refetch: vi.fn(),
}));

vi.mock('@/entities/user', () => ({
  useCurrentUser: mocks.useCurrentUser,
}));

vi.mock('../../model/use-leaderboard', () => ({
  useLeaderboard: mocks.useLeaderboard,
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

const user: User = {
  id: 'user-1',
  email: 'user@example.com',
  verified: true,
  leaderboard: {
    period: { key: 'week', startAt: '2026-08-10', endAt: '2026-08-17' },
    calculatedAt: '2026-08-10T10:00:00Z',
    nextCalculationAt: '2026-08-10T10:10:00Z',
    player: {
      playerId: 'user-1',
      nickname: 'Пользователь',
      position: 2,
      leaves: 800,
      isTop10: true,
    },
  },
};

describe('LeaderboardCard', () => {
  beforeEach(() => {
    mocks.useCurrentUser.mockReset().mockReturnValue({ data: undefined });
    mocks.useLeaderboard.mockReset();
    mocks.refetch.mockReset();
  });

  it('показывает резервную позицию и повторяет запрос без данных', async () => {
    const interaction = userEvent.setup();
    mocks.useLeaderboard.mockReturnValue({
      data: undefined,
      isPending: false,
      refetch: mocks.refetch,
    });
    render(<LeaderboardCard />);

    expect(screen.getByText('Ваше место: 0')).toBeInTheDocument();
    await interaction.click(screen.getByRole('article'));
    expect(mocks.refetch).toHaveBeenCalledOnce();
  });

  it('показывает позицию и список при наличии данных пользователя', async () => {
    const interaction = userEvent.setup();
    mocks.useCurrentUser.mockReturnValue({ data: user });
    mocks.useLeaderboard.mockReturnValue({
      data: {
        items: [{ playerId: 'user-1', nickname: 'Пользователь', position: 2, leaves: 800 }],
      },
      isPending: false,
      refetch: mocks.refetch,
    });
    render(<LeaderboardCard />);

    expect(screen.getByText('Ваше место: 2')).toBeInTheDocument();
    expect(screen.getByText('вы')).toBeInTheDocument();
    await interaction.click(screen.getByRole('article'));
    expect(mocks.refetch).not.toHaveBeenCalled();
  });
});
