import avitoBonusIcon from '@/shared/assets/icon/avito-bonus-icon.svg';
import deliveryDiscountIcon from '@/shared/assets/icon/delivery-discount-icon.svg';
import freeDeliveryIcon from '@/shared/assets/icon/free-delivery-icon.svg';
import freePromotionIcon from '@/shared/assets/icon/free-promotion-icon.svg';
import promotionDiscountIcon from '@/shared/assets/icon/promotion-discount-icon.svg';

import type { TRewardCategory } from '../api/get-rewards';

export const REWARD_CATEGORY_ICONS: Record<TRewardCategory, string> = {
  AVITO_BONUS: avitoBonusIcon,
  FREE_DELIVERY: freeDeliveryIcon,
  FREE_PROMOTION: freePromotionIcon,
  PROMOTION_DISCOUNT: promotionDiscountIcon,
  DELIVERY_DISCOUNT: deliveryDiscountIcon,
} as const;

export type TRewardCategoryIconType = keyof typeof REWARD_CATEGORY_ICONS;
