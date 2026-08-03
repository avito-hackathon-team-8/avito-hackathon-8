import { useEffect, useState } from 'react';

import styles from './main.page.module.scss';

type ServiceStatus = {
  service: string;
  status: string;
  time: string;
};

export const MainPage = () => {
  const [data, setData] = useState<ServiceStatus | null>(null);
  const [error, setError] = useState(false);

  useEffect(() => {
    fetch('/api/app/status')
      .then((response) => {
        if (!response.ok) {
          throw new Error('Backend unavailable');
        }

        return response.json() as Promise<ServiceStatus>;
      })
      .then(setData)
      .catch(() => setError(true));
  }, []);

  return (
    <main className={styles.page}>
      <section className={styles.card} aria-live="polite">
        <span className={styles.eyebrow}>Local starter</span>
        <h1>React meets PocketBase</h1>
        <p className={styles.description}>
          A small full-stack foundation running through Docker Compose.
        </p>
        <div className={styles.status}>
          <span className={error ? styles['error-dot'] : styles.dot} />
          {error && 'Backend unavailable'}
          {!error && !data && 'Connecting to PocketBase'}
          {data && `${data.service} is ${data.status}`}
        </div>
        {data && (
          <time className={styles.time} dateTime={data.time}>
            Checked {new Date(data.time).toLocaleString()}
          </time>
        )}
        <a href="http://localhost:8090/_/">Open PocketBase dashboard</a>
      </section>
    </main>
  );
};
