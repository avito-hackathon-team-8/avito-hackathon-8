import { usePetProfile } from '@/entities/gamification-profile';
import { Button } from '@/shared/ui/button';

import { usePetInteraction } from '../../model/use-pet-interaction';

import styles from './pet-interaction.module.scss';

export const PetInteraction = () => {
  const { data: pet } = usePetProfile();
  const { carePet, isPending } = usePetInteraction();

  return (
    <div className={styles.petInteraction}>
      <Button
        className={styles.petInteraction__buttonFood}
        variant="primary"
        disabled={isPending || pet?.feedNextAvailableAt !== null}
        onClick={() => carePet('FEED')}
      >
        Покормить
      </Button>

      <Button
        className={styles.petInteraction__buttonPet}
        variant="primary"
        disabled={isPending || pet?.strokeNextAvailableAt !== null}
        onClick={() => carePet('STROKE')}
      >
        Погладить
      </Button>
    </div>
  );
};
