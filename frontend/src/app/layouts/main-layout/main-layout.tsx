import { Outlet } from "react-router";

import { Header } from "@/widgets/header";

const test = [1, 2, 3, 4, 5];

export const MainLayout = () => {
  return (
    <main>
      <Header petName="tt" />
      <Outlet />
      {test.map((i) => (
        <span key={i}>{i}</span>
      ))}
    </main>
  );
};
