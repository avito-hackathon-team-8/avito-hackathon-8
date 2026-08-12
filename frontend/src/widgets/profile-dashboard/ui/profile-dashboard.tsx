import { GamificationProfile, GamificationScene } from '@/entities/gamification-profile';
import { BuyReward } from '@/features/buy-reward';
import { PetInteraction } from '@/features/pet-interaction';

import styles from './profile-dashboard.module.scss';

export const ProfileDashboard = () => {
  return (
    <div className={styles.profileDashboard}>
      <GamificationScene />
      <GamificationProfile />
      <PetInteraction />
      <BuyReward />
    </div>
  );
};
