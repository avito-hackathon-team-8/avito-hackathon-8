import { ActivityDaysAutoPanel } from '@/entities/activity-days';
import { usePetName } from '@/entities/gamification-profile/model/use-pet-name';
import { usePetProfileSocket } from '@/entities/gamification-profile/model/use-pet-profile-socket';
import { useTodaySummarySocket } from '@/entities/today-summary';
import { PetNameModal } from '@/features/pet-name';
import { GamificationDashboard } from '@/widgets/gamification-dashboard';
import { ProfileDashboard } from '@/widgets/profile-dashboard/ui/profile-dashboard';

import styles from './main.page.module.scss';

export const MainPage = () => {
  const { data: pet } = usePetName();

  const isPetInitialized = Boolean(pet?.trim());

  usePetProfileSocket({ enabled: isPetInitialized });
  useTodaySummarySocket();

  return (
    <>
      <PetNameModal isOpen={!pet || pet.trim().length === 0} />
      {isPetInitialized && (
        <>
          <div className={styles.page}>
            <ProfileDashboard />
            <GamificationDashboard />
          </div>
          <ActivityDaysAutoPanel />
        </>
      )}
    </>
  );
};
