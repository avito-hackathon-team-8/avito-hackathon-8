import clsx from 'clsx';
import type { HTMLAttributes, ImgHTMLAttributes } from 'react';

import { Typography } from '../typography';

import styles from './gamification-card.module.scss';

interface IGamificationCardProps {
  title: string;
  description: string;
  imageProps: ImgHTMLAttributes<HTMLImageElement>;
  wrapperProps: Omit<HTMLAttributes<HTMLElement>, 'className'>;
  variant?: 'vertical' | 'horizontal';
  className?: string;
}

export const GamificationCard = ({
  title,
  description,
  imageProps,
  wrapperProps,
  variant = 'vertical',
  className,
}: IGamificationCardProps) => {
  return (
    <article className={clsx(styles.card, styles[`card_${variant}`], className)} {...wrapperProps}>
      <Typography as="h3" variant="p3">
        {title}
      </Typography>

      <Typography className={styles.card__description} variant="caption-medium" color="gray500">
        {description}
      </Typography>

      <img className={clsx(styles.card__img, imageProps.className)} {...imageProps} />
    </article>
  );
};
