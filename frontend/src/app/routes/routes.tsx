import { createBrowserRouter } from 'react-router';

import { AppShell } from '@/app/layouts/app-shell/app-shell';
import { MainLayout } from '@/app/layouts/main-layout';
import { TanstackQueryProvider } from '@/app/providers/tanstack-query';
import { MainPage } from '@/pages/main-page/main.page';
import { RegisterPage } from '@/pages/register-page';
import { APP_ROUTES } from '@/shared/config/routes';

import { GuestOnly } from './guards/guest-only';
import { RequireAuth } from './guards/require-auth';

export const browserRouter = createBrowserRouter([
  {
    element: <TanstackQueryProvider />,
    children: [
      {
        element: <AppShell />,
        children: [
          {
            element: <GuestOnly />,
            children: [
              {
                path: APP_ROUTES.auth,
                element: <RegisterPage />,
              },
            ],
          },
          {
            element: <RequireAuth />,
            children: [
              {
                element: <MainLayout />,
                children: [
                  {
                    path: APP_ROUTES.home,
                    element: <MainPage />,
                  },
                ],
              },
            ],
          },
        ],
      },
    ],
  },
]);
