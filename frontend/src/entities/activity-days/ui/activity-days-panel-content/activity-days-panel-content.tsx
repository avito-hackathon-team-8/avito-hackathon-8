import clsx from "clsx";

import { formatDays } from "@/shared/lib/format-days";
import { Typography } from "@/shared/ui/typography";

import type { TResponseActivityDay } from "../../api/get-activity-day";

import styles from "./activity-days-panel-content.module.scss";

interface IActivityDaysPanelContentProps {
  data: TResponseActivityDay;
}

export const ActivityDaysPanelContent = ({
  data,
}: IActivityDaysPanelContentProps) => {
  const { claimedDaysCount, claims } = data;

  const activeDay = claims.find((item) => item.status === "AVAILABLE");
  return (
    <div className={styles.wrapper}>
      <ul className={styles.days}>
        {claims.map(({ weekday, status, rewardLeaves, date }) => (
          <li
            key={date}
            className={clsx(styles.day, styles[`day_${status.toLowerCase()}`])}
          >
            <button
              className={styles.day__button}
              disabled={status !== "AVAILABLE"}
            >
              <time className={styles.day__date} dateTime={date}>
                <Typography
                  className={styles.day__dateText}
                  variant="caption-bold"
                  color="inherit"
                >
                  {weekday}д
                </Typography>
              </time>

              <Typography
                className={styles.day__reward}
                variant="caption-bold"
                color="inherit"
              >
                {rewardLeaves}
              </Typography>
            </button>
          </li>
        ))}
      </ul>

      <div className={styles.info__wrapper}>
        <Typography variant="caption-bold" color="white">
          День {activeDay?.weekday} — {activeDay?.rewardLeaves} листьев
        </Typography>
        <Typography variant="caption" color="white">
          Зайдите сегодня, чтобы получить награду
        </Typography>
      </div>

      {activeDay && (
        <Typography className={styles.info__series} variant="caption-bold">
          Текущая серия: {formatDays(claimedDaysCount)}
        </Typography>
      )}
    </div>
  );
};
