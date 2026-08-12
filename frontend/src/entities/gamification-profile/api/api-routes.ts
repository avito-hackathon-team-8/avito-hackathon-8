export const API_ROUTE_PROFILE = {
  pet: 'v1/pet',
  petProfileWs: '/v1/pet/ws',
  levels: '/v1/pet/levels',
  receiveLevelReward: (rewardId: string) => `/v1/pet/level-rewards/${rewardId}/claim`,
};
