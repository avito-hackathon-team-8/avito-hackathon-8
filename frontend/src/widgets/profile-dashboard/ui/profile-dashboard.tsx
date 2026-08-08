import { GamificationProfile } from '@/entities/gamification-profile';
import { GamificationScene } from '@/entities/gamification-profile/ui/gamification-scene/gamification-scene';
import { BuyReward } from '@/features/buy-reward/ui/buy-reward/buy-reward';

import styles from './profile-dashboard.module.scss';

export const ProfileDashboard = () => {
  return (
    <div className={styles.profileDashboard}>
      <GamificationScene />
      <GamificationProfile />
      <BuyReward />
    </div>
  );
};
