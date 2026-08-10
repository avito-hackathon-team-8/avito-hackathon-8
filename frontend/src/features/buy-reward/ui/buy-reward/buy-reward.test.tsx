import { render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';

import { BuyReward } from './buy-reward';

const mocks = vi.hoisted(() => ({
  useBuyReward: vi.fn(),
  openChest: vi.fn(),
  addMVPLeaves: vi.fn(),
  closeModal: vi.fn(),
}));

vi.mock('../../model/use-buy-reward', () => ({
  useBuyReward: mocks.useBuyReward,
}));

const defaultState = {
  pet: {
    name: 'Листик',
    level: 10,
    leaves: 500,
    nextLevelTargetLeaves: 0,
    chestPrice: 100,
  },
  reward: null,
  isOpen: false,
  isPending: false,
  isDisabled: false,
  isMVPLeavesPending: false,
  isRewardVisible: false,
  openChest: mocks.openChest,
  addMVPLeaves: mocks.addMVPLeaves,
  closeModal: mocks.closeModal,
};

describe('BuyReward', () => {
  beforeEach(() => {
    const portalRoot = document.createElement('div');
    portalRoot.id = 'app-modal-root';
    document.body.append(portalRoot);

    mocks.useBuyReward.mockReset().mockReturnValue(defaultState);
    mocks.openChest.mockReset();
    mocks.addMVPLeaves.mockReset();
    mocks.closeModal.mockReset();
  });

  afterEach(() => {
    document.getElementById('app-modal-root')?.remove();
    document.body.style.overflow = '';
  });

  it('показывает цену и вызывает действия пользователя', async () => {
    const user = userEvent.setup();
    render(<BuyReward />);

    expect(screen.getByText(/Стоимость открытия 100/)).toBeInTheDocument();

    await user.click(screen.getByRole('button', { name: /Открыть сундук/ }));
    await user.click(screen.getByRole('button', { name: 'MVP: +200 листьев' }));

    expect(mocks.openChest).toHaveBeenCalledOnce();
    expect(mocks.addMVPLeaves).toHaveBeenCalledOnce();
  });

  it('отображает pending-состояния и блокирует действия', () => {
    mocks.useBuyReward.mockReturnValue({
      ...defaultState,
      isPending: true,
      isDisabled: true,
      isMVPLeavesPending: true,
    });

    render(<BuyReward />);

    expect(screen.getByRole('button', { name: /Открываем сундук/ })).toBeDisabled();
    expect(screen.getByRole('button', { name: 'Начисляем листья...' })).toBeDisabled();
  });

  it('показывает полученную награду и закрывает модальное окно', async () => {
    const user = userEvent.setup();
    mocks.useBuyReward.mockReturnValue({
      ...defaultState,
      isOpen: true,
      isRewardVisible: true,
      reward: { title: 'Бесплатная доставка' },
    });

    render(<BuyReward />);

    expect(screen.getByRole('dialog')).toHaveTextContent('Бесплатная доставка');
    await user.click(screen.getByRole('button', { name: 'Закрыть' }));
    expect(mocks.closeModal).toHaveBeenCalledOnce();
  });
});
