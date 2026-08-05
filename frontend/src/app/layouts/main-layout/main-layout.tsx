import { Outlet } from "react-router";

import { Header } from "@/widgets/header";

export const MainLayout = () => {
  return (
    <main>
      <Header />
      <Outlet />
    </main>
  );
};
