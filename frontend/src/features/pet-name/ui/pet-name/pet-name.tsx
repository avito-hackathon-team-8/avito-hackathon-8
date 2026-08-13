import { usePetProfile } from '@/entities/gamification-profile';
import { Typography } from '@/shared/ui/typography';

import styles from './pet-name.module.scss';

interface IPetNameProps {
  className?: string;
}

const getStatus = (happiness: number) => {
  if (happiness >= 80) {
    return { text: 'хорошее', color: 'green700' as const };
  }

  if (happiness >= 35) {
    return { text: 'нейтральное', color: 'gray500' as const };
  }

  return { text: 'плохое', color: 'red' as const };
};

export const PetName = ({ className }: IPetNameProps) => {
  const { data: pet } = usePetProfile();

  const happiness = pet?.happiness ?? 0;
  const status = getStatus(happiness);

  return (
    <Typography className={className} as="h1" variant="heading">
      {pet?.name}

      <Typography className={styles.name} as="span" variant="caption-medium">
        Настроение:{' '}
        <Typography
          className={styles.name__status}
          as="span"
          variant="caption-medium"
          color={status.color}
        >
          {status.text} ({happiness.toFixed(1)})
        </Typography>
      </Typography>
    </Typography>
  );
};
