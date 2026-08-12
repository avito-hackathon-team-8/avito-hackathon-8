import { BottomPanel } from '@/shared/ui/bottom-panel';
import { GamificationCard } from '@/shared/ui/gamification-card';

import { useShopItems } from '../../model/use-shop-items';
import basketIcon from '../assets/basket-icon.webp';
import { ShopThingsContent } from '../shop-things-content/shop-things-content';

const TITLE_CARD = 'Магазин аксессуаров';
const DESCRIPTION_TEXT = (length: number) => `Товаров для комнаты: ${length}`;
const DESCRIPTION_TEXT_EMPTY = 'Товаров нет';

type TShopThingsCardProps = {
  className?: string;
};

export const ShopThingsCard = ({ className }: TShopThingsCardProps) => {
  const { data: shopItems, isPending, refetch, purchaseItem, isPurchasePending } = useShopItems();

  return (
    <BottomPanel
      title={TITLE_CARD}
      description=""
      disabled={!shopItems}
      onClick={() => {
        if (!shopItems && !isPending) {
          refetch();
        }
      }}
      renderTrigger={(open) => (
        <GamificationCard
          title={TITLE_CARD}
          description={shopItems ? DESCRIPTION_TEXT(shopItems.length) : DESCRIPTION_TEXT_EMPTY}
          imageProps={{ src: basketIcon, alt: 'Корзина товаров', width: 90, height: 90 }}
          wrapperProps={{ onClick: open }}
          variant="horizontal"
          className={className}
        />
      )}
    >
      {shopItems && (
        <ShopThingsContent
          shopItems={shopItems}
          purchaseItem={purchaseItem}
          isPurchasePending={isPurchasePending}
        />
      )}
    </BottomPanel>
  );
};
