import { useState } from 'react';

import clsx from 'clsx';

import { usePetProfile } from '@/entities/gamification-profile';
import leafIcon from '@/shared/assets/icon/leaf-icon.svg';
import { resolveAssetUrl } from '@/shared/config';
import { Button } from '@/shared/ui/button';
import { Confirm } from '@/shared/ui/confirm';
import { Typography } from '@/shared/ui/typography';

import type { TShopItem, TShopItemCategory } from '../../api/shop-items';
import type { TPurchaseShopItemVariables } from '../../model/use-shop-items';

import styles from './shop-things-content.module.scss';

const SHOP_CATEGORIES: { category: TShopItemCategory; title: string }[] = [
  { category: 'BOWL', title: 'Миска' },
  { category: 'BED', title: 'Лежанка' },
];

const REPLACEMENT_CONFIRM_TITLE =
  'При покупке нового предмета текущий будет удалён. Чтобы вернуть его, потребуется купить его заново.';

type TShopThingsContentProps = {
  shopItems: TShopItem[];
  purchaseItem: (variables: TPurchaseShopItemVariables) => void;
  isPurchasePending: boolean;
};

export const ShopThingsContent = ({
  shopItems,
  purchaseItem,
  isPurchasePending,
}: TShopThingsContentProps) => {
  const { data: pet } = usePetProfile();
  const [activeCategory, setActiveCategory] = useState<TShopItemCategory>('BOWL');
  const [confirmationItemId, setConfirmationItemId] = useState<TShopItem['id'] | null>(null);

  const filteredItems = shopItems.filter(({ category }) => category === activeCategory);
  const hasActiveItem = filteredItems.some(({ status }) => status === 'ACTIVE');

  return (
    <div className={styles.content}>
      <div className={styles.content__header}>
        <menu className={styles.content__filter}>
          {SHOP_CATEGORIES.map(({ category, title }) => (
            <li className={styles.content__filterItem} key={category}>
              <Button
                className={clsx(styles.content__filterButton, {
                  [styles.content__filterButton_active]: activeCategory === category,
                })}
                onClick={() => {
                  setActiveCategory(category);
                  setConfirmationItemId(null);
                }}
              >
                {title}
              </Button>
            </li>
          ))}
        </menu>
      </div>

      <ul className={styles.content__list}>
        {filteredItems.map((item) => {
          const isDisabled = pet ? item.requiredLevel > pet?.level : true;
          const isActive = item.status === 'ACTIVE';
          const isAvailable = item.status === 'AVAILABLE';

          return (
            <li className={styles.content__item} key={item.id}>
              <div className={styles.content__itemContainer}>
                <img src={resolveAssetUrl(item.imageUrl)} width={60} height={60} alt={item.title} />

                <div className={styles.content__itemWrapper}>
                  <Typography variant="p3">{item.title}</Typography>
                  <Typography variant="caption">{item.description}</Typography>
                </div>

                <Button
                  className={clsx(styles.content__button, {
                    [styles.content__button_active]: isActive,
                    [styles.content__button_buy]: isAvailable,
                  })}
                  disabled={!isAvailable || isPurchasePending || isDisabled}
                  onClick={() => {
                    if (hasActiveItem) {
                      setConfirmationItemId(item.id);

                      return;
                    }

                    purchaseItem({
                      itemId: item.id,
                      confirmReplacement: false,
                    });
                  }}
                >
                  <Typography className={styles.content__currency} variant="caption-bold">
                    {isActive && 'Активно'}
                    {item.status === 'LOCKED' && `Разблокируется на ${item.requiredLevel} уровне`}
                    {isAvailable && (
                      <>
                        купить за {item.priceLeaves} <img src={leafIcon} aria-hidden />
                      </>
                    )}
                  </Typography>
                </Button>

                <Typography className={styles.content__description} variant="caption">
                  Срок действия {item.durationDays} дня
                </Typography>
              </div>

              {confirmationItemId === item.id && (
                <Confirm
                  title={REPLACEMENT_CONFIRM_TITLE}
                  disabled={isPurchasePending}
                  onConfirm={() => {
                    purchaseItem({
                      itemId: item.id,
                      confirmReplacement: true,
                    });
                    setConfirmationItemId(null);
                  }}
                  onCancel={() => setConfirmationItemId(null)}
                />
              )}
            </li>
          );
        })}
      </ul>

      {filteredItems.length === 0 && (
        <Typography variant="caption">Товаров в этой категории нет</Typography>
      )}
    </div>
  );
};
