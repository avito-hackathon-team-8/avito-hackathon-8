import { GamificationProfile, GamificationScene } from '@/entities/gamification-profile';
import { BuyReward } from '@/features/buy-reward';
import { PetInteraction } from '@/features/pet-interaction';

import styles from './profile-dashboard.module.scss';

export const ProfileDashboard = () => {
  return (
    <div className={styles.profileDashboard}>
      <div data-tutorial="pet">
        <GamificationScene />
      </div>
      <div data-tutorial="progress">
        <GamificationProfile />
      </div>
      <div data-tutorial="pet-care">
        <PetInteraction />
      </div>
      <BuyReward />
    </div>
  );
};
