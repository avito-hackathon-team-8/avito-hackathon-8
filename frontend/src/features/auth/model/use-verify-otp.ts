import { useMutation, useQueryClient } from '@tanstack/react-query';

import { userQueryKeys } from '@/entities/user';
import { getCurrentUser, verifyOtp } from '@/features/auth/api/auth';
import { mainQueryKey } from '@/shared/config/api';
import { sessionStorageKeysMap, setSessionStorageValue } from '@/shared/lib/session-storage';

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
