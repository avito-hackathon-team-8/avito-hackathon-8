import { ArrowIcon } from "@/shared/assets/icon/arrow";
import { Button } from "@/shared/ui/button";
import { Typography } from "@/shared/ui/typography";

import { ButtonNotification } from "./button-notification/button-notification";

import styles from "./header.module.scss";

interface HeaderProps {
  petName: string;
}

export const Header = ({ petName }: HeaderProps) => {
  return (
    <header className={styles.header}>
      <Button
        className={styles.header__backButton}
        aria-label="Вернуться назад"
      >
        <ArrowIcon></ArrowIcon>
      </Button>

      <Typography className={styles.header__name} as="h1" variant="heading">
        {petName}
      </Typography>

      <ButtonNotification className={styles.header__notificationButton} />
    </header>
  );
};
