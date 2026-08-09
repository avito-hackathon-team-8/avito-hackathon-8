import clsx from 'clsx';

import { formatDays } from '@/shared/lib';
import { Button } from '@/shared/ui/button';
import { Typography } from '@/shared/ui/typography';

import type { TResponseActivityDay } from '../../api/activity-day';

import styles from './activity-days-panel-content.module.scss';

interface IActivityDaysPanelContentProps {
  data: TResponseActivityDay;
  handleReceiveReward: () => void;
}

const LIST_DAYS = {
  1: 'пн',
  2: 'вт',
  3: 'ср',
  4: 'чт',
  5: 'пт',
  6: 'сб',
  7: 'вс',
};

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
        {claims.map(({ weekday, status, rewardLeaves, date }) => (
          <li key={date} className={clsx(styles.day, styles[`day_${status.toLowerCase()}`])}>
            <button
              className={styles.day__button}
              disabled={status !== 'AVAILABLE'}
              onClick={handleReceiveReward}
            >
              <time className={styles.day__date} dateTime={date}>
                <Typography className={styles.day__dateText} variant="caption-bold" color="inherit">
                  {LIST_DAYS[weekday]}
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
          onClick={handleReceiveReward}
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
