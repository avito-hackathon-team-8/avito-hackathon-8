export const gamificationProfileKeys = {
  all: ['gamification-profile'] as const,
  pet: () => [...gamificationProfileKeys.all, 'pet'] as const,
  levels: () => [...gamificationProfileKeys.all, 'levels'] as const,
};
