import { Typography } from '@/shared/ui/typography';

import styles from './today-summary-empty.module.scss';

export const TodaySummaryEmpty = () => {
  return (
    <div className={styles.empty}>
      <Typography as="h3" variant="section">
        Сегодня пока нет активности
      </Typography>

      <Typography as="p" variant="caption" color="gray500">
        Выполняйте задания — результаты появятся в сводке
      </Typography>
    </div>
  );
};
