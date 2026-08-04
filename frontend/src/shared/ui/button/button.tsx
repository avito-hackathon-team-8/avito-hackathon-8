import clsx from "clsx";
import type { ButtonHTMLAttributes, JSX } from "react";

import styles from "./button.module.scss";

type TVariant = "primary" | "transparent";

interface IButtonProps extends ButtonHTMLAttributes<HTMLButtonElement> {
  variant?: TVariant;
  icon?: JSX.Element;
}

export const Button = ({
  className,
  children,
  variant,
  ...props
}: IButtonProps) => {
  return (
    <button
      className={clsx(styles.button, styles[`button__${variant}`], className)}
      {...props}
    >
      {children}
    </button>
  );
};
