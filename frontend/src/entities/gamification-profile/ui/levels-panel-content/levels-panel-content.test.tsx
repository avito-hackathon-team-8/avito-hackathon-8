import { render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { describe, expect, it, vi } from 'vitest';

import type { TLevelRewardItem, TLevelRewardStatus } from '../../api/levels-rewards';

import { LevelsPanelContent } from './levels-panel-content';

const createLevel = (level: number, status: TLevelRewardStatus): TLevelRewardItem => ({
  level,
  status,
  reward: {
    id: `reward-${level}`,
    type: 'FREE_DELIVERY',
    description: `Награда уровня ${level}`,
  },
  expiresAt: status === 'UNOPENED' ? '2026-08-20T10:00:00Z' : null,
});

describe('LevelsPanelContent', () => {
  it('показывает состояния уровней и выдаёт доступную награду', async () => {
    const user = userEvent.setup();
    const handleReceiveReward = vi.fn();
    const levels = [
      createLevel(1, 'CLAIMED'),
      createLevel(2, 'FROZEN'),
      createLevel(3, 'UNOPENED'),
      createLevel(4, 'LOCKED'),
    ];

    render(<LevelsPanelContent levelsList={levels} handleReceiveReward={handleReceiveReward} />);

    expect(screen.getByText('Награда получена')).toBeInTheDocument();
    expect(screen.getByText('Награда пропущена')).toBeInTheDocument();
    expect(screen.getByText('2026-08-20')).toBeInTheDocument();
    expect(screen.getByText('Разблокируется на 4 уровне')).toBeInTheDocument();

    const buttons = screen.getAllByRole('button', { name: 'Забрать награду' });
    expect(buttons.filter((button) => !button.hasAttribute('disabled'))).toHaveLength(1);
    await user.click(buttons[2]);
    expect(handleReceiveReward).toHaveBeenCalledWith({ rewardId: 'reward-3', level: 3 });
  });
});
