import { usePetName } from '@/entities/gamification-profile/model/use-pet-name';
import { usePetProfileSocket } from '@/entities/gamification-profile/model/use-pet-profile-socket';
import { PetNameModal } from '@/features/pet-name';
import { GamificationDashboard } from '@/widgets/gamification-dashboard';
import { ProfileDashboard } from '@/widgets/profile-dashboard/ui/profile-dashboard';

import styles from './main.page.module.scss';

export const MainPage = () => {
  const { data: pet } = usePetName();

  usePetProfileSocket({ enabled: Boolean(pet?.length === 0) });

  return (
    <>
      <PetNameModal isOpen={!pet || pet.trim().length === 0} />
      {pet && pet.length > 0 && (
        <div className={styles.page}>
          <ProfileDashboard />
          <GamificationDashboard />
        </div>
      )}
    </>
  );
};
