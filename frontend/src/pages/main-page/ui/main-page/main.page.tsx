import { ActivityDaysAutoPanel } from '@/entities/activity-days';
import { usePetName, usePetProfileSocket } from '@/entities/gamification-profile';
import { useTodaySummarySocket } from '@/entities/today-summary';
import { useBasicTutorial } from '@/features/basic-tutorial';
import { PetNameModal } from '@/features/pet-name';
import { GamificationDashboard } from '@/widgets/gamification-dashboard';
import { ProfileDashboard } from '@/widgets/profile-dashboard';

import { WelcomeOverlay } from '../welcome-overlay/welcome-overlay';

import styles from './main.page.module.scss';

export const MainPage = () => {
  const { data: pet } = usePetName();

  const isPetInitialized = Boolean(pet?.trim());
  const { startTutorial } = useBasicTutorial({ enabled: isPetInitialized });

  usePetProfileSocket({ enabled: isPetInitialized });
  useTodaySummarySocket();

  return (
    <>
      <PetNameModal isOpen={!pet || pet.trim().length === 0} />
      {isPetInitialized && (
        <>
          <WelcomeOverlay />
          <div className={styles.page}>
            <ProfileDashboard />
            <GamificationDashboard onStartTutorial={startTutorial} />
          </div>
          <ActivityDaysAutoPanel />
        </>
      )}
    </>
  );
};
