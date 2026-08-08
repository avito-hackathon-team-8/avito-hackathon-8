export const userQueryKeys = {
  all: ['user'] as const,
  current: () => [...userQueryKeys.all, 'current'] as const,
};
