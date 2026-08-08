import { useEffect } from 'react';

import { Outlet, useLocation, useNavigate } from 'react-router';

import { ToasterProvider } from '@/app/providers/toaster';
import { useCurrentUser } from '@/entities/user';
import { APP_ROUTES } from '@/shared/config/routes';
import { Typography } from '@/shared/ui/typography';

import styles from './app-shell.module.scss';

export const AppShell = () => {
  const navigate = useNavigate();
  const location = useLocation();

  const { data: currentUser, isLoading } = useCurrentUser();

  useEffect(() => {
    if (isLoading) {
      return;
    }

    if (!currentUser && location.pathname !== APP_ROUTES.auth) {
      navigate(APP_ROUTES.auth, { replace: true });
    }
  }, [currentUser, isLoading, location.pathname, navigate]);

  return (
    <div className={styles.appShell}>
      <div className={styles.appShell__container}>
        {isLoading ? <Typography variant="heading">Загрузка</Typography> : <Outlet />}

        <div id="app-overlay-root" className={styles.appShell__overlayRoot} />
        <div id="app-modal-root" />
        <ToasterProvider />
      </div>
    </div>
  );
};
