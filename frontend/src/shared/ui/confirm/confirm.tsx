import { Button } from '../button';
import { Typography } from '../typography';

import styles from './confirm.module.scss';

type TConfirmProps = {
  title: string;
  onConfirm: () => void;
  onCancel: () => void;
  disabled?: boolean;
};

export const Confirm = ({ title, onConfirm, onCancel, disabled = false }: TConfirmProps) => {
  return (
    <div className={styles.confirm} role="alert">
      <Typography className={styles.confirm__title} variant="caption">
        {title}
      </Typography>

      <div className={styles.confirm__actions}>
        <Button
          className={styles.confirm__button}
          variant="primary"
          aria-label="Подтвердить"
          disabled={disabled}
          onClick={onConfirm}
        >
          ✓
        </Button>

        <Button
          className={styles.confirm__button}
          variant="default"
          aria-label="Отменить"
          disabled={disabled}
          onClick={onCancel}
        >
          ×
        </Button>
      </div>
    </div>
  );
};

export type { TConfirmProps };
