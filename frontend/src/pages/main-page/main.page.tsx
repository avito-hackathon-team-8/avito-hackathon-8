import { GamificationDashboard } from '@/widgets/gamification-dashboard';
import { ProfileDashboard } from '@/widgets/profile-dashboard/ui/profile-dashboard';

import styles from './main.page.module.scss';

export const MainPage = () => {
  return (
    <div className={styles.page}>
      <ProfileDashboard />
      <GamificationDashboard />
    </div>
  );
};
