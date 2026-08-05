import { useMutation, useQueryClient } from "@tanstack/react-query";

import { userQueryKeys } from "@/entities/user";
import { verifyOtp } from "@/features/auth/api/auth";
import {
  sessionStorageKeysMap,
  setSessionStorageValue,
} from "@/shared/lib/session-storage";

import { authMutationKeys } from "./auth-query-keys";

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

    onSuccess: ({ token, record }) => {
      setSessionStorageValue(sessionStorageKeysMap.authToken, token);

      queryClient.setQueryData(userQueryKeys.current(), record);
    },
  });
};
