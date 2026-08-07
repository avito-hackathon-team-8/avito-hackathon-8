import { usePetProfile } from '@/entities/gamification-profile/model/use-pet-profile';
import { PetNameModal } from '@/features/pet-name';
import { Typography } from '@/shared/ui/typography';
import { GamificationDashboard } from '@/widgets/gamification-dashboard';
import { ProfileDashboard } from '@/widgets/profile-dashboard/ui/profile-dashboard';

import styles from './main.page.module.scss';

export const MainPage = () => {
  const { data: pet, isLoading, isError } = usePetProfile();

  if (isLoading) {
    return (
      <div className={styles.page}>
        <Typography as="p" variant="body" color="gray500">
          Загрузка...
        </Typography>
      </div>
    );
  }

  if (isError || !pet) {
    return (
      <div className={styles.page}>
        <Typography as="p" variant="body" color="red">
          Не удалось загрузить данные питомца.
        </Typography>
      </div>
    );
  }

  return (
    <>
      <PetNameModal isOpen={!pet.name.trim()} />
      <div className={styles.page}>
        <ProfileDashboard />
        <GamificationDashboard />
      </div>
    </>
  );
};
