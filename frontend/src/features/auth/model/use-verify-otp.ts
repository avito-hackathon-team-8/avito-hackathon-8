import { useMutation, useQueryClient } from '@tanstack/react-query';

import { getCurrentUser, userQueryKeys } from '@/entities/user';
import { mainQueryKey } from '@/shared/config';
import { sessionStorageKeysMap, setSessionStorageValue } from '@/shared/lib';

import { verifyOtp } from '../api/auth';

import { authMutationKeys } from './auth-query-keys';

type VerifyOtpParams = {
  email: string;
  code: string;
};

export const useVerifyOtp = () => {
  const queryClient = useQueryClient();

  return useMutation({
    mutationKey: authMutationKeys.verifyOtp(),

    mutationFn: ({ email, code }: VerifyOtpParams) => {
      return verifyOtp(email, code);
    },

    onSuccess: async ({ token }) => {
      setSessionStorageValue(sessionStorageKeysMap.authToken, token);

      await queryClient.fetchQuery({
        queryKey: userQueryKeys.current(),
        queryFn: () => getCurrentUser(token),
      });

      await queryClient.invalidateQueries({
        queryKey: mainQueryKey.all,
      });
    },
  });
};
