import clsx from 'clsx';
import type { AnchorHTMLAttributes, ButtonHTMLAttributes, JSX } from 'react';

import styles from './button.module.scss';

type TVariant = 'primary' | 'transparent' | 'default';

type TBaseButtonProps = {
  variant?: TVariant;
  icon?: JSX.Element;
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
    const { as: Component, className, children, variant, icon, ...anchorProps } = props;

    return (
      <Component
        className={clsx(styles.button, styles[`button__${variant}`], className)}
        {...anchorProps}
      >
        {icon}
        {children}
      </Component>
    );
  }

  const { as: Component = 'button', className, children, variant, icon, ...buttonProps } = props;

  return (
    <Component
      className={clsx(styles.button, styles[`button__${variant}`], className)}
      {...buttonProps}
    >
      {icon}
      {children}
    </Component>
  );
};
