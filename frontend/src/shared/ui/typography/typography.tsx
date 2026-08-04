import {
  type ComponentPropsWithoutRef,
  type ElementType,
  type ReactNode,
  type Ref,
} from "react";

import clsx from "clsx";

import styles from "./Typography.module.scss";

export type TypographyTag =
  | "h1"
  | "h2"
  | "h3"
  | "h4"
  | "p"
  | "span"
  | "div"
  | "a"
  | "label"
  | "legend"
  | "caption"
  | "strong"
  | "small";

export type TypographyVariant =
  "inherit" | "display" | "heading" | "section" | "body" | "caption";

export type TypographyColor =
  | "inherit"
  | "white"
  | "gray-100"
  | "gray-200"
  | "gray-500"
  | "black"
  | "blue"
  | "green"
  | "red"
  | "purple";

interface TypographyOwnProps<T extends TypographyTag> {
  as: T;
  variant: TypographyVariant;
  color?: TypographyColor;
  className?: string;
  children?: ReactNode;
  ref?: Ref<HTMLElementTagNameMap[T]>;
}

export type TypographyProps<T extends TypographyTag> = TypographyOwnProps<T> &
  Omit<ComponentPropsWithoutRef<T>, keyof TypographyOwnProps<T> | "color">;

const variantClassNames: Record<TypographyVariant, string> = {
  inherit: styles.inherit,
  display: styles.display,
  heading: styles.heading,
  section: styles.section,
  body: styles.body,
  caption: styles.caption,
};

const colorClassNames: Record<TypographyColor, string> = {
  inherit: styles.colorInherit,
  white: styles.colorWhite,
  "gray-100": styles.colorGray100,
  "gray-200": styles.colorGray200,
  "gray-500": styles.colorGray500,
  black: styles.colorBlack,
  blue: styles.colorBlue,
  green: styles.colorGreen,
  red: styles.colorRed,
  purple: styles.colorPurple,
};

function Typography<T extends TypographyTag>({
  as,
  variant,
  color = "black",
  className,
  children,
  ref,
  ...props
}: TypographyProps<T>) {
  const Component = as as ElementType;

  return (
    <Component
      ref={ref}
      className={clsx(
        styles.root,
        variantClassNames[variant],
        colorClassNames[color],
        className,
      )}
      {...props}
    >
      {children}
    </Component>
  );
}

export { Typography };
