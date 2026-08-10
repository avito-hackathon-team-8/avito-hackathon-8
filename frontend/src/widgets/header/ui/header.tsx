import { PetName } from '@/features/pet-name';
import { ArrowIcon } from '@/shared/assets/icon';
import { Button } from '@/shared/ui/button';

import styles from './header.module.scss';

export const Header = () => {
  return (
    <header className={styles.header}>
      <Button
        as="a"
        href="https://www.avito.ru/profile"
        className={styles.header__backButton}
        aria-label="Вернуться назад"
      >
        <ArrowIcon></ArrowIcon>
      </Button>

      <PetName className={styles.header__name} />
    </header>
  );
};
