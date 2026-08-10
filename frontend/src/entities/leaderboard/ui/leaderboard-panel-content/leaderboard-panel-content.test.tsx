import { render, screen } from '@testing-library/react';
import { describe, expect, it } from 'vitest';

import type { User } from '@/entities/user';

import type { TLeaderboardUser } from '../../api/leaderboard';

import { LeaderboardPanelContent } from './leaderboard-panel-content';

const createUser = (position: number): User => ({
  id: 'current-user',
  email: 'user@example.com',
  verified: true,
  leaderboard: {
    period: {
      key: 'week',
      startAt: '2026-08-10T00:00:00Z',
      endAt: '2026-08-17T00:00:00Z',
    },
    calculatedAt: '2026-08-10T10:00:00Z',
    nextCalculationAt: '2026-08-10T10:10:00Z',
    player: {
      playerId: 'current-user',
      nickname: 'Пользователь',
      position,
      leaves: 350,
      isTop10: position <= 10,
    },
  },
});

const leaders: TLeaderboardUser[] = [
  { playerId: 'leader-1', nickname: 'Анна', position: 1, leaves: 900 },
  { playerId: 'current-user', nickname: 'Пользователь', position: 2, leaves: 800 },
];

describe('LeaderboardPanelContent', () => {
  it('помечает текущего пользователя внутри топа', () => {
    render(<LeaderboardPanelContent listUsers={leaders} userPosition={2} user={createUser(2)} />);

    expect(screen.getByText('Анна')).toBeInTheDocument();
    expect(screen.getByText('вы')).toBeInTheDocument();
    expect(screen.queryByText('•••')).not.toBeInTheDocument();
  });

  it('добавляет позицию пользователя за пределами топа', () => {
    render(
      <LeaderboardPanelContent
        listUsers={leaders.slice(0, 1)}
        userPosition={42}
        user={createUser(42)}
      />,
    );

    expect(screen.getByText('•••')).toBeInTheDocument();
    expect(screen.getByText('Вы')).toBeInTheDocument();
    expect(screen.getByText('42')).toBeInTheDocument();
    expect(screen.getByText('350')).toBeInTheDocument();
  });
});
