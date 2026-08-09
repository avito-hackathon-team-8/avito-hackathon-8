import clsx from 'clsx';

import { usePetProfile } from '@/entities/gamification-profile';
import { Typography } from '@/shared/ui/typography';

interface IPetNameProps {
  className?: string;
}

export const PetName = ({ className }: IPetNameProps) => {
  const { data } = usePetProfile();

  return (
    <Typography className={clsx(className)} as="h1" variant="heading">
      {data?.name}
    </Typography>
  );
};
