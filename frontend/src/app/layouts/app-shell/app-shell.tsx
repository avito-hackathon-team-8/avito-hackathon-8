import { Outlet } from "react-router";

import styles from "./app-shell.module.scss";

export const AppShell = () => {
  return (
    <main className={styles.appShell}>
      <div className={styles.appShell__container}>
        <Outlet />
        <div id="app-overlay-root" className={styles.appShell__overlayRoot} />
      </div>
    </main>
  );
};
