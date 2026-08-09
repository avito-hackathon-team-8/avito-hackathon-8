import { type SubmitEvent, useState } from 'react';

import { useMutation, useQueryClient } from '@tanstack/react-query';
import { toast } from 'sonner';

import {
  gamificationProfileKeys,
  updatePetName,
  usePetProfile,
} from '@/entities/gamification-profile';
import { Button } from '@/shared/ui/button';
import { Modal } from '@/shared/ui/modal';
import { Typography } from '@/shared/ui/typography';

import styles from './pet-name-modal.module.scss';

type PetNameModalProps = {
  isOpen: boolean;
};

const TEXT_ERROR = 'Не удалось сохранить имя питомца';

export const PetNameModal = ({ isOpen }: PetNameModalProps) => {
  const queryClient = useQueryClient();
  const { data: pet } = usePetProfile();
  const [name, setName] = useState('');
  const [error, setError] = useState('');

  const mutation = useMutation({
    mutationFn: updatePetName,
    onSuccess: (updatedPet) => {
      queryClient.setQueryData(gamificationProfileKeys.pet(), updatedPet);
      setName('');
      setError('');
    },
    onError: (submitError) => {
      toast.error(TEXT_ERROR);
      setError(submitError instanceof Error ? submitError.message : TEXT_ERROR);
    },
  });

  const handleSubmit = (event: SubmitEvent<HTMLFormElement>) => {
    event.preventDefault();

    const nextName = name.trim();

    if (nextName.length < 1 || nextName.length > 20) {
      setError('Имя должно содержать от 1 до 20 символов');
      return;
    }

    setError('');

    mutation.mutate(nextName);
  };

  const shouldOpen = Boolean(isOpen && pet && !pet.name.trim());

  return (
    <Modal isOpen={shouldOpen} onClose={() => undefined}>
      <div className={styles.modal}>
        <div className={styles.modal__header}>
          <Typography as="h2" variant="section">
            Давайте назовём питомца
          </Typography>
          <Typography as="p" variant="caption" color="gray500">
            Придумайте имя от 1 до 20 символов
          </Typography>
        </div>

        <form className={styles.modal__form} onSubmit={handleSubmit} noValidate>
          <label className={styles.modal__field}>
            <Typography as="span" variant="caption" color="gray500">
              Имя питомца
            </Typography>
            <input
              className={styles.modal__input}
              type="text"
              name="pet-name"
              autoFocus
              maxLength={20}
              value={name}
              placeholder="Например: Листик"
              aria-invalid={Boolean(error)}
              onChange={(event) => {
                setName(event.target.value);
                if (error) {
                  setError('');
                }
              }}
            />
          </label>

          {error ? (
            <Typography as="p" variant="caption" color="red" className={styles.modal__error}>
              {error}
            </Typography>
          ) : null}

          <Button
            type="submit"
            variant="primary"
            className={styles.modal__submit}
            disabled={mutation.isPending || !name.trim()}
          >
            {mutation.isPending ? 'Сохраняем...' : 'Сохранить'}
          </Button>
        </form>
      </div>
    </Modal>
  );
};
