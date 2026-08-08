import { useQuery } from '@tanstack/react-query';

import { getCurrentUser } from '@/features/auth/api/auth';
import { getSessionStorageValue, sessionStorageKeysMap } from '@/shared/lib/session-storage';

import { userQueryKeys } from '../api/user-query-keys';

export const useCurrentUser = () => {
  const token = getSessionStorageValue(sessionStorageKeysMap.authToken);

  return useQuery({
    queryKey: userQueryKeys.current(),
    queryFn: () => getCurrentUser(token),
    enabled: Boolean(token),
    retry: false,
    staleTime: 30_000,
  });
};
