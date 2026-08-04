import { Outlet } from "react-router";

const test = [1, 2, 3, 4, 5];

export const MainLayout = () => {
  return (
    <main>
      <Outlet />
      {test.map((i) => (
        <span key={i}>{i}</span>
      ))}
    </main>
  );
};
