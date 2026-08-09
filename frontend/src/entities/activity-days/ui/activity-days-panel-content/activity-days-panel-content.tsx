import clsx from 'clsx';

import { formatDays } from '@/shared/lib/format-days';
import { Button } from '@/shared/ui/button';
import { Typography } from '@/shared/ui/typography';

import type { TActivityDay, TResponseActivityDay } from '../../api/activity-day';

import styles from './activity-days-panel-content.module.scss';

interface IActivityDaysPanelContentProps {
  data: TResponseActivityDay;
  handleReceiveReward: (claimId: NonNullable<TActivityDay['claimId']>) => void;
}

export const ActivityDaysPanelContent = ({
  data,
  handleReceiveReward,
}: IActivityDaysPanelContentProps) => {
  const { claimedDaysCount, claims } = data;
  const firstFutureIndex = claims.findIndex(({ status }) => status === 'FUTURE');

  const activeDay = firstFutureIndex === -1 ? claims.at(-1) : claims[firstFutureIndex - 1];

  return (
    <div className={styles.wrapper}>
      <ul className={styles.days}>
        {claims.map(({ weekday, status, rewardLeaves, date, claimId }) => (
          <li key={date} className={clsx(styles.day, styles[`day_${status.toLowerCase()}`])}>
            <button
              className={styles.day__button}
              disabled={status !== 'AVAILABLE'}
              onClick={() => handleReceiveReward(claimId)}
            >
              <time className={styles.day__date} dateTime={date}>
                <Typography className={styles.day__dateText} variant="caption-bold" color="inherit">
                  {weekday}д
                </Typography>
              </time>

              <Typography className={styles.day__reward} variant="caption-bold" color="inherit">
                {rewardLeaves}
              </Typography>
            </button>
          </li>
        ))}
      </ul>

      {activeDay && (
        <Button
          className={styles.button}
          disabled={activeDay.status !== 'AVAILABLE'}
          variant="primary"
          isFullWidth
          onClick={() => handleReceiveReward(activeDay.claimId)}
        >
          Забрать награду
        </Button>
      )}

      <Typography className={styles.info__series} variant="caption-bold">
        Текущая серия: {formatDays(claimedDaysCount)}
      </Typography>
    </div>
  );
};
