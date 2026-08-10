import { useEffect, useState } from 'react';

import { usePageScrollLock } from '@/shared/lib';
import { Typography } from '@/shared/ui/typography';

import styles from './welcome-overlay.module.scss';

const DISPLAY_DURATION_MS = 2000;

export const WelcomeOverlay = () => {
  const [isVisible, setIsVisible] = useState(true);

  usePageScrollLock(isVisible);

  useEffect(() => {
    const timeoutId = window.setTimeout(() => {
      setIsVisible(false);
    }, DISPLAY_DURATION_MS);

    return () => window.clearTimeout(timeoutId);
  }, []);

  if (!isVisible) {
    return null;
  }

  return (
    <div className={styles.overlay} role="status" aria-live="polite">
      <Typography as="h1" variant="display" color="blue" className={styles.title}>
        Добро пожаловать!
      </Typography>
    </div>
  );
};
