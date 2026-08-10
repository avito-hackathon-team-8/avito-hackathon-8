import { useNavigate } from 'react-router';

import { AuthFormStep, useOtpAuth, WelcomeStep } from '@/features/auth';
import { APP_ROUTES } from '@/shared/config';

import styles from './register-page.module.scss';

export const RegisterPage = () => {
  const navigate = useNavigate();

  const auth = useOtpAuth({
    onSuccess: () => {
      navigate(APP_ROUTES.home, { replace: true });
    },
  });

  return (
    <div className={styles.page}>
      {auth.step === 'welcome' ? <WelcomeStep onContinue={auth.openEmailStep} /> : null}

      {auth.step === 'email' ? (
        <AuthFormStep
          variant="email"
          value={auth.email}
          error={auth.error}
          isSubmitting={auth.isRequestingCode}
          isValid={auth.isEmailValid}
          onChange={auth.changeEmail}
          onSubmit={auth.sendCode}
        />
      ) : null}

      {auth.step === 'code' ? (
        <AuthFormStep
          variant="code"
          value={auth.code}
          error={auth.error}
          isSubmitting={auth.isVerifyingCode}
          isValid={auth.isCodeValid}
          onChange={auth.changeCode}
          onBack={auth.returnToEmailStep}
          onSubmit={auth.signInByCode}
        />
      ) : null}
    </div>
  );
};
