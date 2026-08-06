import { GamificationDashboard } from '@/widgets/gamification-dashboard';

import styles from './main.page.module.scss';

export const MainPage = () => {
  return (
    <div className={styles.wrapper}>
      <GamificationDashboard />
    </div>
  );
};
