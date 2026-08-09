import { Button } from '@/shared/ui/button';
import { Typography } from '@/shared/ui/typography';

import type { TReward } from '../../api/rewards';
import { RewardsIcons } from '../assets/reward-icon';

import styles from './rewards-panel-content.module.scss';

interface IRewardsPanelContentProps {
  listReward: TReward[];
}

export const RewardsPanelContent = ({ listReward }: IRewardsPanelContentProps) => {
  return (
    <>
      {listReward.length === 0 && (
        <div className={styles.rewardsEmpty}>
          <Typography variant="body">
            Повышайте уровень и&nbsp;открывайте сундуки —&nbsp;награды появятся здесь.
          </Typography>
        </div>
      )}
      <div className={styles.rewardsPanel}>
        <ul className={styles.rewardsPanel__list}>
          {listReward.map((data) => (
            <li key={data.id} className={styles.rewardsPanel__item}>
              <RewardsIcons
                className={styles.rewardsPanel__itemImg}
                aria-hidden
                variant={data.category}
              />

              <div className={styles.rewardsPanel__itemContent}>
                <Typography variant="p2-semiBold" as="h3">
                  {data.categoryName}
                </Typography>

                <Typography variant="p4-regular" color="gray500">
                  до <time dateTime={data.expiresAt}>{data.expiresAt.split('T')[0]}</time>
                </Typography>
              </div>

              <Button
                as="a"
                target="_blank"
                href="https://www.avito.ru/"
                className={styles.rewardsPanel__itemButton}
                variant="default"
              >
                <Typography variant="p4-bold">Применить</Typography>
              </Button>
            </li>
          ))}
        </ul>
      </div>
    </>
  );
};
