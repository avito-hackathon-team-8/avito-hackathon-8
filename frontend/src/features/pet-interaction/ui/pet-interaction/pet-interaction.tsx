import { usePetProfile } from '@/entities/gamification-profile';
import { Button } from '@/shared/ui/button';
import { Typography } from '@/shared/ui/typography';

import { usePetInteraction } from '../../model/use-pet-interaction';
import petHappyImg from '../assets/pet-happy.png';
import petNeutralImg from '../assets/pet-neutral.png';
import petSadImg from '../assets/pet-sad.png';

import styles from './pet-interaction.module.scss';

const SETTINGS_PET = {
  happy: {
    src: petHappyImg,
    alt: 'Счастливое лицо питомца',
  },
  neutral: {
    src: petNeutralImg,
    alt: 'Нейтральное лицо питомца',
  },
  sad: {
    src: petSadImg,
    alt: 'Грустное лицо питомца',
  },
};

const getPetSetting = (happiness: number) => {
  if (happiness >= 80) {
    return SETTINGS_PET.happy;
  }

  if (happiness >= 35) {
    return SETTINGS_PET.neutral;
  }

  return SETTINGS_PET.sad;
};

export const PetInteraction = () => {
  const { data: pet } = usePetProfile();
  const { carePet, isPending } = usePetInteraction();
  const happiness = pet?.happiness ?? 0;
  const activeSetting = getPetSetting(happiness);

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

      <div className={styles.petInteraction__mood}>
        <img
          className={styles.petInteraction__moodImg}
          src={activeSetting.src}
          alt={activeSetting.alt}
          width={35}
          height={35}
        />
        <Typography variant="caption-semiBold" color="green700">
          {happiness}
        </Typography>
      </div>

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
