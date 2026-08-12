import { screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { beforeEach, describe, expect, it, vi } from 'vitest';

import type { BottomPanelProps } from '@/shared/ui/bottom-panel/bottom-panel';
import { renderWithProviders } from '@/test/render-with-providers';

import type { TShopItem } from '../../api/shop-items';

import { ShopThingsCard } from './shop-things-card';

const mocks = vi.hoisted(() => ({
  useShopItems: vi.fn(),
  purchaseItem: vi.fn(),
  refetch: vi.fn(),
}));

vi.mock('../../model/use-shop-items', () => ({
  useShopItems: mocks.useShopItems,
}));

type BottomPanelStubProps = Pick<
  BottomPanelProps,
  'children' | 'disabled' | 'onClick' | 'renderTrigger'
>;

vi.mock('@/shared/ui/bottom-panel', () => ({
  BottomPanel: ({ children, disabled, onClick, renderTrigger }: BottomPanelStubProps) => (
    <div data-testid="bottom-panel" data-disabled={disabled}>
      {renderTrigger(() => onClick?.())}
      {children}
    </div>
  ),
}));

const shopItem: TShopItem = {
  id: 'fashionable-bowl',
  category: 'BOWL',
  status: 'AVAILABLE',
  title: 'Модная миска',
  description: 'Кэшбек за покупки одежды.',
  imageUrl: '/api/v1/shop-images/bowl-fashionable.webp',
  requiredLevel: 5,
  priceLeaves: 100,
  durationDays: 3,
};

describe('ShopThingsCard', () => {
  beforeEach(() => {
    mocks.useShopItems.mockReset();
    mocks.purchaseItem.mockReset();
    mocks.refetch.mockReset();
  });

  it('блокирует панель и повторяет запрос по клику без данных', async () => {
    const user = userEvent.setup();
    mocks.useShopItems.mockReturnValue({
      data: undefined,
      isPending: false,
      refetch: mocks.refetch,
      purchaseItem: mocks.purchaseItem,
      isPurchasePending: false,
    });

    renderWithProviders(<ShopThingsCard />);

    expect(screen.getByText('Товаров нет')).toBeInTheDocument();
    expect(screen.getByTestId('bottom-panel')).toHaveAttribute('data-disabled', 'true');

    await user.click(screen.getByRole('article'));

    expect(mocks.refetch).toHaveBeenCalledOnce();
  });

  it('не повторяет запрос во время загрузки', async () => {
    const user = userEvent.setup();
    mocks.useShopItems.mockReturnValue({
      data: undefined,
      isPending: true,
      refetch: mocks.refetch,
      purchaseItem: mocks.purchaseItem,
      isPurchasePending: false,
    });

    renderWithProviders(<ShopThingsCard />);
    await user.click(screen.getByRole('article'));

    expect(mocks.refetch).not.toHaveBeenCalled();
  });

  it('разблокирует панель и передаёт данные в контент', () => {
    mocks.useShopItems.mockReturnValue({
      data: [shopItem],
      isPending: false,
      refetch: mocks.refetch,
      purchaseItem: mocks.purchaseItem,
      isPurchasePending: false,
    });

    renderWithProviders(<ShopThingsCard />);

    expect(screen.getByText('Товаров для комнаты: 1')).toBeInTheDocument();
    expect(screen.getByText('Модная миска')).toBeInTheDocument();
    expect(screen.getByTestId('bottom-panel')).toHaveAttribute('data-disabled', 'false');
  });
});
