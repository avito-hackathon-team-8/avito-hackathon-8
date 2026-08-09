import { useMutation } from '@tanstack/react-query';

import { requestOtp } from '../api/auth';

import { authMutationKeys } from './auth-query-keys';

export const useRequestOtp = () => {
  return useMutation({
    mutationKey: authMutationKeys.requestOtp(),
    mutationFn: (email: string) => requestOtp(email),
  });
};
