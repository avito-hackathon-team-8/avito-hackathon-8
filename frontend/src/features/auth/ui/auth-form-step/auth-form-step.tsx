import type { SubmitEvent } from 'react';

import { Button } from '@/shared/ui/button';
import { Typography } from '@/shared/ui/typography';

import { OTP_LENGTH } from '../../model/constants';

import styles from './auth-form-step.module.scss';

type BaseAuthFormStepProps = {
  value: string;
  error: string;
  isSubmitting: boolean;
  isValid: boolean;
  onChange: (value: string) => void;
  onSubmit: () => void | Promise<void>;
};

type EmailAuthFormStepProps = BaseAuthFormStepProps & {
  variant: 'email';
};

type CodeAuthFormStepProps = BaseAuthFormStepProps & {
  variant: 'code';
  onBack: () => void;
};

type AuthFormStepProps = EmailAuthFormStepProps | CodeAuthFormStepProps;

export const AuthFormStep = (props: AuthFormStepProps) => {
  const isEmailStep = props.variant === 'email';

  const errorId = `${props.variant}-step-error`;

  const handleSubmit = (event: SubmitEvent<HTMLFormElement>) => {
    event.preventDefault();

    void props.onSubmit();
  };

  return (
    <section className={styles.authFormStep} data-variant={props.variant}>
      <div className={styles.authFormStep__header}>
        <Typography as="h1" variant="heading">
          {isEmailStep ? 'Вход' : 'Код из письма'}
        </Typography>

        <Typography as="p" variant="caption" color="gray500">
          {isEmailStep ? 'Введите вашу почту' : `Введите любой 8 значный код`}
        </Typography>
      </div>

      <form className={styles.authFormStep__form} onSubmit={handleSubmit} noValidate>
        <label className={styles.authFormStep__field}>
          <Typography as="span" variant="caption" color="gray500">
            {isEmailStep ? 'Email' : 'Код из письма'}
          </Typography>

          <input
            className={styles.authFormStep__input}
            type={isEmailStep ? 'email' : 'text'}
            name={isEmailStep ? 'email' : 'otp'}
            inputMode={isEmailStep ? 'email' : 'numeric'}
            autoComplete={isEmailStep ? 'email' : 'one-time-code'}
            autoFocus
            maxLength={isEmailStep ? undefined : OTP_LENGTH}
            value={props.value}
            placeholder={isEmailStep ? 'example@mail.ru' : '12345678'}
            disabled={props.isSubmitting}
            aria-invalid={Boolean(props.error)}
            aria-describedby={props.error ? errorId : undefined}
            onChange={(event) => {
              props.onChange(event.target.value);
            }}
          />
        </label>

        {props.error ? (
          <Typography
            id={errorId}
            as="p"
            variant="caption"
            color="red"
            className={styles.authFormStep__error}
          >
            {props.error}
          </Typography>
        ) : null}

        <div className={styles.authFormStep__actions}>
          <Button
            type="submit"
            variant="primary"
            className={styles.authFormStep__submitButton}
            disabled={props.isSubmitting || !props.isValid}
          >
            {isEmailStep
              ? props.isSubmitting
                ? 'Отправляем...'
                : 'Получить код'
              : props.isSubmitting
                ? 'Проверяем...'
                : 'Войти'}
          </Button>

          {!isEmailStep ? (
            <Button
              type="button"
              variant="primary"
              className={styles.authFormStep__backButton}
              aria-label="Вернуться к вводу почты"
              disabled={props.isSubmitting}
              onClick={props.onBack}
            >
              Вернуться назад
            </Button>
          ) : null}
        </div>
      </form>
    </section>
  );
};
