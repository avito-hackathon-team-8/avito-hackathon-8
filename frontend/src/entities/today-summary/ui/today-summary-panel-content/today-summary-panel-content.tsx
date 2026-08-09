import { DailyTaskIcons } from '@/entities/daily-task/ui/assets/daily-task-icons';
import { RewardsIcons } from '@/entities/reward/ui/assets/reward-icon';
import leafIcon from '@/shared/assets/icon/leaf-icon.svg';
import { LevelUpIcon } from '@/shared/assets/icon/level-up-icon';
import { formatTasks } from '@/shared/lib/format-tasks';
import { formatWord } from '@/shared/lib/format-word';
import { Typography } from '@/shared/ui/typography';

import type { TTodaySummaryStats } from '../../model/get-today-summary-stats ';

import styles from './today-summary-panel-content.module.scss';

interface ITodaySummaryPanelContentProps {
  events: TTodaySummaryStats;
}

const completedWords = ['выполненное', 'выполненных', 'выполненных'] as const;

export const TodaySummaryPanelContent = ({ events }: ITodaySummaryPanelContentProps) => {
  return (
    <div className={styles.panel}>
      <div className={styles.panel__header}>
        <Typography variant="p2-semiBold">
          {events.activitiesCount} {formatWord(events.activitiesCount, completedWords)}{' '}
          {formatTasks(events.activitiesCount)}
        </Typography>

        <div className={styles.panel__headerCount}>
          <Typography variant="p2-bold" color="green700">
            + {events.leavesCount}
          </Typography>
          <img src={leafIcon} width={24} height={24} aria-hidden />
        </div>
      </div>

      <ul className={styles.panel__list}>
        {events.events.map((item) => {
          if (item.eventType === 'REWARD') {
            const { type, title, expiresAt } = item.data;
            return (
              <li className={styles.panel__item} key={item.data.rewardId}>
                <RewardsIcons className={styles.panel__icon} variant={type} aria-hidden />

                <Typography className={styles.panel__title} variant="p2-semiBold" as="h3">
                  {title}
                </Typography>
                <Typography
                  className={styles.panel__description}
                  variant="p4-regular"
                  color="gray500"
                >
                  Можно получить до <time dateTime={expiresAt}>{expiresAt.split('T')[0]}</time>
                </Typography>
              </li>
            );
          }

          if (item.eventType === 'LEVEL_UP') {
            const { occurredAt, fromLevel, toLevel } = item.data;
            return (
              <li className={styles.panel__item} key={occurredAt}>
                <LevelUpIcon className={styles.panel__icon} />

                <Typography className={styles.panel__title} variant="p2-semiBold" as="h3">
                  Уровень повышен с {fromLevel} до {toLevel}
                </Typography>
                <Typography
                  className={styles.panel__description}
                  variant="p4-regular"
                  color="gray500"
                >
                  Время получения <time dateTime={occurredAt}>{occurredAt.split('T')[0]}</time>
                </Typography>
              </li>
            );
          }
          if (item.eventType === 'TASK') {
            const { description, rewardLeaves, taskId } = item.data;

            return (
              <li className={styles.panel__item} key={taskId}>
                <DailyTaskIcons
                  className={styles.panel__icon}
                  variant={item.data.type}
                  aria-hidden
                />

                <Typography className={styles.panel__title} variant="p2-semiBold" as="h3">
                  {description}
                </Typography>
                <Typography
                  className={styles.panel__description}
                  variant="p4-regular"
                  color="gray500"
                >
                  Полвучено листьев{' '}
                  <Typography variant="p4-regular" color="green700" as="span">
                    {rewardLeaves}
                  </Typography>
                </Typography>
              </li>
            );
          }
        })}
      </ul>
    </div>
  );
};
