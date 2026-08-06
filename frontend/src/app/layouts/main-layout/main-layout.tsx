import { Outlet } from 'react-router';

import { Header } from '@/widgets/header';

import styles from './main-layout.module.scss';

export const MainLayout = () => {
  return (
    <>
      <Header />
      <main className={styles.page__main}>
        <Outlet />
      </main>
    </>
  );
};
