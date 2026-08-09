import { APP_ROUTES } from '@/shared/config';
import { Button } from '@/shared/ui/button';
import { Typography } from '@/shared/ui/typography';

import pageNotFoundImage from './assets/page-not-found.svg';

import styles from './page-not-found.module.scss';

export const PageNotFound = () => {
  return (
    <main className={styles.page}>
      <div className={styles.page__content}>
        <div className={styles.page__copy}>
          <Typography as="h1" variant="heading">
            Такой страницы не существует
          </Typography>

          <Typography as="p" variant="caption" color="gray500">
            Она могла быть удалена или никогда не существовала.
          </Typography>

          <Button as="a" href={APP_ROUTES.home} variant="primary" className={styles.page__button}>
            Вернуться на главную
          </Button>
        </div>

        <img
          className={styles.page__image}
          src={pageNotFoundImage}
          alt="Персонаж ищет пропавшую страницу"
          width={1254}
          height={1254}
        />
      </div>
    </main>
  );
};
