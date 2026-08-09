import { Outlet } from 'react-router';

import { usePetName } from '@/entities/gamification-profile';
import { AppError } from '@/shared/ui/app-error';
import { AppLoader } from '@/shared/ui/app-loader';
import { Header } from '@/widgets/header';

import styles from './main-layout.module.scss';

export const MainLayout = () => {
  const { data: petName, isLoading, isError } = usePetName();

  const isInitialized = Boolean(petName?.trim());

  if (isLoading) {
    return <AppLoader />;
  }

  if (isError) {
    return <AppError />;
  }

  return (
    <>
      {isInitialized && <Header />}

      <main className={styles.page__main}>
        <Outlet />
      </main>
    </>
  );
};
