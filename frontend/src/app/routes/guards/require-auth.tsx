import { Navigate, Outlet, useLocation } from 'react-router';

import { useCurrentUser } from '@/entities/user';
import { APP_ROUTES } from '@/shared/config';
import { Typography } from '@/shared/ui/typography';

export const RequireAuth = () => {
  const location = useLocation();
  const { data: currentUser, isLoading } = useCurrentUser();

  if (isLoading) {
    return <Typography variant="heading">Загрузка</Typography>;
  }

  if (!currentUser) {
    return <Navigate to={APP_ROUTES.auth} replace state={{ from: location }} />;
  }

  return <Outlet />;
};
