import clsx from 'clsx';
import type { AnchorHTMLAttributes, ButtonHTMLAttributes, JSX } from 'react';

import styles from './button.module.scss';

type TVariant = 'primary' | 'transparent' | 'default';

type TBaseButtonProps = {
  variant?: TVariant;
  icon?: JSX.Element;
  isFullWidth?: boolean;
};

type TButtonElementProps = TBaseButtonProps &
  ButtonHTMLAttributes<HTMLButtonElement> & {
    as?: 'button';
  };

type TAnchorElementProps = TBaseButtonProps &
  AnchorHTMLAttributes<HTMLAnchorElement> & {
    as: 'a';
  };

type TButtonProps = TButtonElementProps | TAnchorElementProps;

export const Button = (props: TButtonProps) => {
  if (props.as === 'a') {
    const {
      as: Component,
      className,
      children,
      variant,
      icon,
      isFullWidth,
      ...anchorProps
    } = props;

    return (
      <Component
        className={clsx(
          styles.button,
          styles[`button_${variant}`],
          { [styles.button_fullWidth]: isFullWidth },
          className,
        )}
        {...anchorProps}
      >
        {icon}
        {children}
      </Component>
    );
  }

  const {
    as: Component = 'button',
    className,
    children,
    variant,
    isFullWidth,
    icon,
    ...buttonProps
  } = props;

  return (
    <Component
      className={clsx(
        styles.button,
        styles[`button_${variant}`],
        { [styles.button_fullWidth]: isFullWidth },
        className,
      )}
      {...buttonProps}
    >
      {icon}
      {children}
    </Component>
  );
};
