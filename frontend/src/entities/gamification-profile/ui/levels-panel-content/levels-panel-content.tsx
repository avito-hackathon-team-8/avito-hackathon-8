import clsx from 'clsx';

import { Typography } from '@/shared/ui/typography';

import type { TLevelRewardItem } from '../../api/levels-rewards';
import type { TReceiveRewardVariables } from '../../model/use-levels-profile';

import styles from './levels-panel-content.module.scss';

interface IGamificationProfilePanelContentProps {
  classname?: string;
  levelsList: TLevelRewardItem[];
  handleReceiveReward: (id: TReceiveRewardVariables) => void;
}

export const LevelsPanelContent = ({
  classname,
  levelsList,
  handleReceiveReward,
}: IGamificationProfilePanelContentProps) => {
  return (
    <ul className={clsx(styles.panel, classname)}>
      {levelsList.map(({ level, status, reward, expiresAt }) => {
        const isClaimed = status === 'CLAIMED';
        const isFrozen = status === 'FROZEN';
        const isUnopened = status === 'UNOPENED';
        const isLocked = status === 'LOCKED';

        return (
          <li
            className={clsx(styles.panel__item, {
              [styles.panel__item_claimed]: isClaimed,
              [styles.panel__item_frozen]: isFrozen,
              [styles.panel__item_unopened]: isUnopened,
              [styles.panel__item_locked]: isLocked,
            })}
            key={level}
          >
            <Typography className={styles.panel__count} variant="p3" as="p" color="inherit">
              {!isLocked && level}
              {isLocked && '×'}
            </Typography>

            <Typography
              className={styles.panel__title}
              variant="p2-semiBold"
              as="h3"
              color="inherit"
            >
              {reward.description}
            </Typography>

            <Typography className={styles.panel__description} variant="p4-regular" as="p">
              {isClaimed && 'Награда получена'}
              {isFrozen && 'Награда пропущена'}
              {isUnopened && (
                <>
                  <span>Награду можно забрать до </span>

                  {expiresAt && <time dateTime={expiresAt}>{expiresAt.split('T')[0]}</time>}
                </>
              )}

              {isLocked && `Разблокируется на ${level} уровне`}
            </Typography>

            <button
              aria-label="Забрать награду"
              className={styles.panel__button}
              disabled={!isUnopened}
              onClick={
                isUnopened ? () => handleReceiveReward({ rewardId: reward.id, level }) : undefined
              }
            />
          </li>
        );
      })}
    </ul>
  );
};
