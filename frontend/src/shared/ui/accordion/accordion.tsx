import { type ComponentProps, type PropsWithChildren, useId, useState } from 'react';

import clsx from 'clsx';

import { Typography } from '../typography';

import styles from './accordion.module.scss';

type TAccordionProps = PropsWithChildren<{
  title: string;
  defaultOpen?: boolean;
  titleProps?: ComponentProps<typeof Typography>;
  classNameBody?: string;
}>;

export const Accordion = ({
  title,
  children,
  defaultOpen = false,
  classNameBody,
  titleProps = {
    variant: 'caption-semiBold',
  },
}: TAccordionProps) => {
  const [isOpen, setIsOpen] = useState(defaultOpen);
  const contentId = useId();

  const handleToggle = () => {
    setIsOpen((previousValue) => !previousValue);
  };

  return (
    <div className={styles.accordion} data-open={isOpen}>
      <button
        className={styles.accordion__trigger}
        type="button"
        aria-expanded={isOpen}
        aria-controls={contentId}
        onClick={handleToggle}
      >
        <Typography {...titleProps}>{title}</Typography>

        <svg
          className={styles.accordion__icon}
          width="20"
          height="20"
          viewBox="0 0 20 20"
          fill="none"
          aria-hidden="true"
        >
          <path
            d="M5 7.5L10 12.5L15 7.5"
            stroke="currentColor"
            strokeWidth="1.5"
            strokeLinecap="round"
            strokeLinejoin="round"
          />
        </svg>
      </button>

      <div id={contentId} className={styles.accordion__content} aria-hidden={!isOpen}>
        <div className={styles.accordion__contentInner}>
          <div className={clsx(styles.accordion__body, classNameBody)}>{children}</div>
        </div>
      </div>
    </div>
  );
};
