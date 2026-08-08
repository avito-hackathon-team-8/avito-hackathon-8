export const authMutationKeys = {
  all: ['auth'] as const,

  requestOtp: () => [...authMutationKeys.all, 'request-otp'] as const,

  verifyOtp: () => [...authMutationKeys.all, 'verify-otp'] as const,
};
