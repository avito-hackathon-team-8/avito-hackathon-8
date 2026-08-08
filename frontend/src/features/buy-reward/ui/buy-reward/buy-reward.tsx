import clsx from 'clsx';

import { usePetProfile } from '@/entities/gamification-profile/model/use-pet-profile';
import { Button } from '@/shared/ui/button';
import { Typography } from '@/shared/ui/typography';

import styles from './buy-reward.module.scss';

interface IBuyRewardProps {
  className?: string;
}

export const BuyReward = ({ className }: IBuyRewardProps) => {
  const { data } = usePetProfile();

  const isDisabled = !data || data.leaves < data.chestPrice;

  return (
    <Button className={clsx(styles.buttonBuy, className)} variant="primary" disabled={isDisabled}>
      <Typography variant="p3-semiBold" as="span" color="inherit">
        Открыть сундук
      </Typography>

      <Typography variant="caption" as="span" color="inherit">
        Разблокируется на 10 уровне
      </Typography>
    </Button>
  );
};
