import { usePetProfile } from '@/entities/gamification-profile';

import { PetNameModal } from '../pet-name-modal/pet-name-modal';

export const PetNameModalController = () => {
  const { data: pet } = usePetProfile();

  if (pet === undefined) {
    return null;
  }

  return <PetNameModal isOpen={pet.name.trim().length === 0} />;
};
