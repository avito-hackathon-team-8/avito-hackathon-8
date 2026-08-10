import { useState } from 'react';

import { type AuthStep, EMAIL_PATTERN, OTP_LENGTH } from './constants';
import { useRequestOtp } from './use-request-otp';
import { useVerifyOtp } from './use-verify-otp';

type UseOtpAuthParams = {
  onSuccess?: () => void;
};

const getMutationErrorMessage = (error: unknown, fallbackMessage: string) => {
  return error instanceof Error ? error.message : fallbackMessage;
};

export const useOtpAuth = ({ onSuccess }: UseOtpAuthParams = {}) => {
  const requestOtpMutation = useRequestOtp();
  const verifyOtpMutation = useVerifyOtp();

  const [step, setStep] = useState<AuthStep>('welcome');
  const [email, setEmail] = useState('');
  const [code, setCode] = useState('');
  const [error, setError] = useState('');

  const normalizedEmail = email.trim();
  const normalizedCode = code.trim();

  const isEmailValid = EMAIL_PATTERN.test(normalizedEmail);
  const isCodeValid = normalizedCode.length === OTP_LENGTH;

  const isRequestingCode = requestOtpMutation.isPending;
  const isVerifyingCode = verifyOtpMutation.isPending;
  const isPending = isRequestingCode || isVerifyingCode;

  const openEmailStep = () => {
    setError('');
    setStep('email');
  };

  const returnToEmailStep = () => {
    setCode('');
    setError('');
    setStep('email');
  };

  const changeEmail = (value: string) => {
    setEmail(value);
    setError('');
  };

  const changeCode = (value: string) => {
    const normalizedValue = value.replace(/\D/g, '').slice(0, OTP_LENGTH);

    setCode(normalizedValue);
    setError('');
  };

  const sendCode = async () => {
    if (!isEmailValid) {
      setError('Введите корректный email');

      return;
    }

    setError('');

    try {
      await requestOtpMutation.mutateAsync(normalizedEmail);
      setStep('code');
    } catch (requestError) {
      setError(getMutationErrorMessage(requestError, 'Не удалось отправить код'));
    }
  };

  const signInByCode = async () => {
    if (!isCodeValid) {
      setError(`Введите ${OTP_LENGTH}-значный код`);

      return;
    }

    setError('');

    try {
      await verifyOtpMutation.mutateAsync({
        email: normalizedEmail,
        code: normalizedCode,
      });

      onSuccess?.();
    } catch (verifyError) {
      setError(getMutationErrorMessage(verifyError, 'Не удалось подтвердить код'));
    }
  };

  return {
    step,
    email,
    code,
    error,

    isEmailValid,
    isCodeValid,
    isPending,
    isRequestingCode,
    isVerifyingCode,

    openEmailStep,
    returnToEmailStep,
    changeEmail,
    changeCode,
    sendCode,
    signInByCode,
  };
};
