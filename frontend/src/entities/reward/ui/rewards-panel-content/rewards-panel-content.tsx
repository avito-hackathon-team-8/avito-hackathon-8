import { Button } from '@/shared/ui/button';
import { Typography } from '@/shared/ui/typography';

import type { TReward } from '../../api/get-rewards';
import { REWARD_CATEGORY_ICONS } from '../../model/icons';

import styles from './rewards-panel-content.module.scss';

interface IRewardsPanelContentProps {
  listReward: TReward[];
}

export const RewardsPanelContent = ({ listReward }: IRewardsPanelContentProps) => {
  return (
    <div className={styles.rewardsPanel}>
      <ul className={styles.rewardsPanel__list}>
        {listReward.map((data) => (
          <li key={data.id} className={styles.rewardsPanel__item}>
            <img
              className={styles.rewardsPanel__itemImg}
              src={REWARD_CATEGORY_ICONS[data.category]}
              aria-hidden
            />

            <div className={styles.rewardsPanel__itemContent}>
              <Typography variant="p2-semiBold" as="h3">
                {data.categoryName}
              </Typography>

              <Typography variant="p4-regular" color="gray500">
                до <time dateTime={data.expiresAt}>{data.expiresAt.split('T')[0]}</time>
              </Typography>
            </div>

            <Button className={styles.rewardsPanel__itemButton} variant="default">
              <Typography variant="p4-bold">Применить</Typography>
            </Button>
          </li>
        ))}
      </ul>
    </div>
  );
};
