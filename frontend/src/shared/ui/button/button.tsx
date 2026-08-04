import clsx from "clsx";
import type { ButtonHTMLAttributes } from "react";

import styles from "./button.module.scss";

type TVariant = "primary";

interface IButtonProps extends ButtonHTMLAttributes<HTMLButtonElement> {
  variant: TVariant;
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
