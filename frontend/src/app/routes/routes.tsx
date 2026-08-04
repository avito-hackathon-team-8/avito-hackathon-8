import { createBrowserRouter } from "react-router";

import { MainLayout } from "@/app/layouts/main-layout";
import { MainPage } from "@/pages/main-page/main.page";

export const browserRouter = createBrowserRouter([
  {
    element: <MainLayout />,
    children: [
      {
        path: "/",
        element: <MainPage />,
      },
    ],
  },
]);
