import { createBrowserRouter } from "react-router";

import { MainLayout } from "@/app/layouts/main-layout";
import { MainPage } from "@/pages/main-page/main.page";

import { AppShell } from "../layouts/app-shell/app-shell";
import { TanstackQueryProvider } from "../providers/tanstack-query.provider";

export const browserRouter = createBrowserRouter([
  {
    element: <TanstackQueryProvider />,
    children: [
      {
        element: <AppShell />,
        children: [
          {
            element: <MainLayout />,
            children: [
              {
                path: "/",
                element: <MainPage />,
              },
            ],
          },
        ],
      },
    ],
  },
]);
