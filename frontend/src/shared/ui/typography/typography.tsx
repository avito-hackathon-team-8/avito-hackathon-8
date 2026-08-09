import { type ComponentPropsWithoutRef, type ElementType, type ReactNode, type Ref } from 'react';

import clsx from 'clsx';

import styles from './typography.module.scss';

export type TypographyTag =
  | 'h1'
  | 'h2'
  | 'h3'
  | 'h4'
  | 'p'
  | 'span'
  | 'div'
  | 'a'
  | 'label'
  | 'legend'
  | 'caption'
  | 'strong'
  | 'small';

export type TypographyVariant =
  | 'inherit'
  | 'display'
  | 'heading'
  | 'heading-semiBold'
  | 'section'
  | 'body'
  | 'caption'
  | 'caption-medium'
  | 'caption-semiBold'
  | 'caption-bold'
  | 'p2-bold'
  | 'p2-semiBold'
  | 'p3'
  | 'p3-semiBold'
  | 'p4-bold'
  | 'p4-regular';

export type TypographyColor =
  | 'inherit'
  | 'white'
  | 'gray100'
  | 'gray200'
  | 'gray250'
  | 'gray500'
  | 'black'
  | 'blue'
  | 'green'
  | 'green400'
  | 'green700'
  | 'red'
  | 'purple';

interface TypographyOwnProps<T extends TypographyTag> {
  as?: T;
  variant: TypographyVariant;
  color?: TypographyColor;
  className?: string;
  children?: ReactNode;
  ref?: Ref<HTMLElementTagNameMap[T]>;
}

export type TypographyProps<T extends TypographyTag> = TypographyOwnProps<T> &
  Omit<ComponentPropsWithoutRef<T>, keyof TypographyOwnProps<T> | 'color'>;

function Typography<T extends TypographyTag>({
  as,
  variant,
  color = 'black',
  className,
  children,
  ref,
  ...props
}: TypographyProps<T>) {
  const Component = (as as ElementType) || 'p';

  return (
    <Component
      ref={ref}
      className={clsx(
        styles.typography,
        styles[`typography_${variant}`],
        styles[`typography_color_${color}`],
        className,
      )}
      {...props}
    >
      {children}
    </Component>
  );
}

export { Typography };
