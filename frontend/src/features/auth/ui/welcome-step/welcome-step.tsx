import { Button } from "@/shared/ui/button";
import { Typography } from "@/shared/ui/typography";

import styles from "./welcome-step.module.scss";

type WelcomeStepProps = {
  onContinue: () => void;
};

export const WelcomeStep = ({ onContinue }: WelcomeStepProps) => {
  return (
    <section className={styles.welcomeStep}>
      <div className={styles.welcomeStep__content}>
        <Typography as="h1" variant="heading">
          Привет, это игра «Коробыш»
        </Typography>

        <Typography as="p" variant="caption" color="gray500">
          Войди, чтобы начать играть
        </Typography>
      </div>

      <Button
        type="button"
        variant="primary"
        className={styles.welcomeStep__button}
        onClick={onContinue}
      >
        Начать
      </Button>
    </section>
  );
};
