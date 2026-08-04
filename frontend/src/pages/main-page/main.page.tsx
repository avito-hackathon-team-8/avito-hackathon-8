import { useEffect, useRef, useState } from 'react';

import type { FormEvent } from 'react';

import {
  logout,
  requestOtp,
  restoreUser,
  type User,
  verifyOtp,
} from '@/shared/api/auth';

import styles from './main.page.module.scss';

type Step = 'email' | 'otp';

export const MainPage = () => {
  const [step, setStep] = useState<Step>('email');
  const [email, setEmail] = useState('');
  const [code, setCode] = useState('');
  const [user, setUser] = useState<User | null>(null);
  const [isRestoring, setIsRestoring] = useState(true);
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [error, setError] = useState('');
  const codeInputRef = useRef<HTMLInputElement>(null);

  useEffect(() => {
    restoreUser()
      .then(setUser)
      .finally(() => setIsRestoring(false));
  }, []);

  useEffect(() => {
    if (step === 'otp') {
      codeInputRef.current?.focus();
    }
  }, [step]);

  const sendCode = async (event: FormEvent<HTMLFormElement>) => {
    event.preventDefault();
    setError('');
    setIsSubmitting(true);

    try {
      await requestOtp(email.trim().toLowerCase());
      setStep('otp');
    } catch (requestError) {
      setError(
        requestError instanceof Error
          ? requestError.message
          : 'Could not send the code.',
      );
    } finally {
      setIsSubmitting(false);
    }
  };

  const confirmCode = async (event: FormEvent<HTMLFormElement>) => {
    event.preventDefault();
    setError('');
    setIsSubmitting(true);

    try {
      setUser(await verifyOtp(email.trim().toLowerCase(), code));
    } catch (requestError) {
      setError(
        requestError instanceof Error
          ? requestError.message
          : 'The code is invalid or expired.',
      );
    } finally {
      setIsSubmitting(false);
    }
  };

  const changeEmail = () => {
    setStep('email');
    setCode('');
    setError('');
  };

  const signOut = () => {
    logout();
    setUser(null);
    changeEmail();
  };

  if (isRestoring) {
    return (
      <main className={styles.page}>
        <p className={styles.loading} aria-live="polite">
          Restoring your session…
        </p>
      </main>
    );
  }

  if (user) {
    return (
      <main className={styles.page}>
        <section className={styles.card}>
          <span className={styles.eyebrow}>Signed in</span>
          <h1>Welcome back</h1>
          <p className={styles.description}>
            You’re authenticated as <strong>{user.email}</strong>.
          </p>
          <div className={styles['success-badge']}>
            <span className={styles.dot} aria-hidden="true" />
            Email verified
          </div>
          <button className={styles.secondary} type="button" onClick={signOut}>
            Sign out
          </button>
        </section>
      </main>
    );
  }

  return (
    <main className={styles.page}>
      <section className={styles.card}>
        <span className={styles.eyebrow}>Passwordless access</span>
        <h1>{step === 'email' ? 'Sign in' : 'Check your inbox'}</h1>
        <p className={styles.description}>
          {step === 'email'
            ? 'Enter your email and we’ll send you a one-time code. New accounts are created automatically.'
            : `We sent an 8-digit code to ${email.trim().toLowerCase()}.`}
        </p>

        {step === 'email' ? (
          <form className={styles.form} onSubmit={sendCode}>
            <label htmlFor="email">Email address</label>
            <input
              id="email"
              name="email"
              type="email"
              autoComplete="email"
              placeholder="you@example.com"
              value={email}
              onChange={(event) => setEmail(event.target.value)}
              required
            />
            <button type="submit" disabled={isSubmitting}>
              {isSubmitting ? 'Sending…' : 'Send sign-in code'}
            </button>
          </form>
        ) : (
          <form className={styles.form} onSubmit={confirmCode}>
            <label htmlFor="code">One-time code</label>
            <input
              ref={codeInputRef}
              id="code"
              name="code"
              type="text"
              inputMode="numeric"
              autoComplete="one-time-code"
              pattern="[0-9]{8}"
              maxLength={8}
              placeholder="00000000"
              value={code}
              onChange={(event) =>
                setCode(event.target.value.replace(/\D/g, '').slice(0, 8))
              }
              required
            />
            <button type="submit" disabled={isSubmitting || code.length !== 8}>
              {isSubmitting ? 'Verifying…' : 'Verify and sign in'}
            </button>
            <button
              className={styles.secondary}
              type="button"
              onClick={changeEmail}
              disabled={isSubmitting}
            >
              Use another email
            </button>
          </form>
        )}

        {error && (
          <p className={styles.error} role="alert">
            {error}
          </p>
        )}
      </section>
    </main>
  );
};
