import { render, screen } from '@testing-library/react';
import { describe, expect, it } from 'vitest';

import type { TReward } from '../../api/rewards';

import { RewardsPanelContent } from './rewards-panel-content';

const reward: TReward = {
  id: 'reward-1',
  title: 'Бесплатная доставка',
  category: 'FREE_DELIVERY',
  categoryName: 'Доставка',
  source: 'CHEST',
  active: true,
  status: 'ACTIVE',
  expiresAt: '2026-08-20T10:00:00Z',
  awardedAt: '2026-08-10T10:00:00Z',
  redeemedAt: null,
};

describe('RewardsPanelContent', () => {
  it('показывает подсказку для пустого списка', () => {
    render(<RewardsPanelContent listReward={[]} />);

    expect(screen.getByText(/Повышайте уровень и открывайте сундуки/)).toBeInTheDocument();
  });

  it('показывает награду, срок действия и ссылку применения', () => {
    render(<RewardsPanelContent listReward={[reward]} />);

    expect(screen.getByRole('heading', { name: 'Доставка' })).toBeInTheDocument();
    expect(screen.getByText('2026-08-20')).toHaveAttribute('datetime', '2026-08-20T10:00:00Z');
    expect(screen.getByRole('link', { name: 'Применить' })).toHaveAttribute(
      'href',
      'https://www.avito.ru/',
    );
  });
});
