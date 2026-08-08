import { Typography } from '@/shared/ui/typography';

import { Button } from '../button';

import styles from './app-error.module.scss';

type TAppErrorProps = {
  title?: string;
  description?: string;
};

export const AppError = ({
  title = 'Что-то пошло не так',
  description = 'Не удалось загрузить данные. Попробуйте обновить страницу.',
}: TAppErrorProps) => {
  return (
    <div className={styles.wrapper}>
      <div className={styles.content}>
        <Typography as="h1" variant="p3">
          {title}
        </Typography>

        <Typography className={styles.description} variant="body" color="gray500">
          {description}
        </Typography>

        <Button
          variant="default"
          className={styles.button}
          type="button"
          onClick={() => window.location.reload()}
        >
          Попробовать снова
        </Button>
      </div>
    </div>
  );
};
