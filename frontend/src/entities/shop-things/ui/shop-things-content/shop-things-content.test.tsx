import { render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { beforeEach, describe, expect, it, vi } from 'vitest';

import type { TShopItem } from '../../api/shop-items';

import { ShopThingsContent } from './shop-things-content';

const mocks = vi.hoisted(() => ({
  purchaseItem: vi.fn(),
  usePetProfile: vi.fn(),
}));

vi.mock('@/entities/gamification-profile', () => ({
  usePetProfile: mocks.usePetProfile,
}));

const shopItems: TShopItem[] = [
  {
    id: 'fashionable-bowl',
    category: 'BOWL',
    status: 'ACTIVE',
    title: 'Модная миска',
    description: 'Кэшбек за покупки одежды.',
    imageUrl: '/api/v1/shop-images/bowl-fashionable.webp',
    requiredLevel: 5,
    priceLeaves: 100,
    durationDays: 3,
  },
  {
    id: 'cyber-bowl',
    category: 'BOWL',
    status: 'AVAILABLE',
    title: 'Кибермиска',
    description: 'Кэшбек за покупки электроники.',
    imageUrl: '/api/v1/shop-images/bowl-cyberpunk.webp',
    requiredLevel: 7,
    priceLeaves: 150,
    durationDays: 3,
  },
  {
    id: 'helper-bowl',
    category: 'BOWL',
    status: 'LOCKED',
    title: 'Миска помощника',
    description: 'Описание заблокированной миски.',
    imageUrl: '/api/v1/shop-images/bowl-helper.webp',
    requiredLevel: 10,
    priceLeaves: 200,
    durationDays: 3,
  },
  {
    id: 'trader-bed',
    category: 'BED',
    status: 'AVAILABLE',
    title: 'Лежанка торговца',
    description: 'Скидка на комиссию.',
    imageUrl: '/api/v1/shop-images/bed-sell.webp',
    requiredLevel: 5,
    priceLeaves: 300,
    durationDays: 3,
  },
];

describe('ShopThingsContent', () => {
  beforeEach(() => {
    mocks.purchaseItem.mockReset();
    mocks.usePetProfile.mockReset().mockReturnValue({ data: { level: 7 } });
  });

  const renderContent = () =>
    render(
      <ShopThingsContent
        shopItems={shopItems}
        purchaseItem={mocks.purchaseItem}
        isPurchasePending={false}
      />,
    );

  it('фильтрует товары по категории', async () => {
    const user = userEvent.setup();
    renderContent();

    expect(screen.getByText('Модная миска')).toBeInTheDocument();
    expect(screen.getByRole('img', { name: 'Модная миска' })).toHaveAttribute(
      'src',
      expect.stringMatching(
        /^http:\/\/localhost:\d+\/api\/v1\/shop-images\/bowl-fashionable\.webp$/,
      ),
    );
    expect(screen.queryByText('Лежанка торговца')).not.toBeInTheDocument();

    await user.click(screen.getByRole('button', { name: 'Лежанка' }));

    expect(screen.getByText('Лежанка торговца')).toBeInTheDocument();
    expect(screen.queryByText('Модная миска')).not.toBeInTheDocument();
  });

  it('предупреждает перед заменой активного товара и подтверждает покупку', async () => {
    const user = userEvent.setup();
    renderContent();

    expect(screen.getByRole('button', { name: 'Активно' })).toBeDisabled();
    expect(screen.getByRole('button', { name: 'Разблокируется на 10 уровне' })).toBeDisabled();

    await user.click(screen.getByRole('button', { name: /купить за 150/ }));

    expect(mocks.purchaseItem).not.toHaveBeenCalled();
    expect(screen.getByRole('alert')).toHaveTextContent(
      'При покупке нового предмета текущий будет удалён',
    );

    await user.click(screen.getByRole('button', { name: 'Подтвердить' }));

    expect(mocks.purchaseItem).toHaveBeenCalledWith({
      itemId: 'cyber-bowl',
      confirmReplacement: true,
    });
    expect(screen.queryByRole('alert')).not.toBeInTheDocument();
  });

  it('отменяет замену активного товара', async () => {
    const user = userEvent.setup();
    renderContent();

    await user.click(screen.getByRole('button', { name: /купить за 150/ }));
    await user.click(screen.getByRole('button', { name: 'Отменить' }));

    expect(mocks.purchaseItem).not.toHaveBeenCalled();
    expect(screen.queryByRole('alert')).not.toBeInTheDocument();
  });

  it('покупает товар сразу, если активного предмета в категории нет', async () => {
    const user = userEvent.setup();
    renderContent();

    await user.click(screen.getByRole('button', { name: 'Лежанка' }));
    await user.click(screen.getByRole('button', { name: /купить за 300/ }));

    expect(mocks.purchaseItem).toHaveBeenCalledWith({
      itemId: 'trader-bed',
      confirmReplacement: false,
    });
    expect(screen.queryByRole('alert')).not.toBeInTheDocument();
  });
});
