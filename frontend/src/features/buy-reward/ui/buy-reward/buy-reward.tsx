import clsx from 'clsx';

import { Button } from '@/shared/ui/button';
import { Typography } from '@/shared/ui/typography';

import styles from './buy-reward.module.scss';

interface IBuyRewardProps {
  className?: string;
}

export const BuyReward = ({ className }: IBuyRewardProps) => {
  return (
    <Button className={clsx(styles.buttonBuy, className)} variant="primary" disabled>
      <Typography variant="p3-semiBold" as="span" color="inherit">
        Открыть сундук
      </Typography>

      <Typography variant="caption" as="span" color="inherit">
        Разблокируется на 10 уровне
      </Typography>
    </Button>
  );
};
