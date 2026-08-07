export const gamificationProfileKeys = {
  all: ['gamification-profile'] as const,
  levels: () => [...gamificationProfileKeys.all, 'levels'] as const,
};
