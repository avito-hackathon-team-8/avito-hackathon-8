import type { TPet } from '@/entities/gamification-profile';
import { apiRequest, getAuthHeaders } from '@/shared/api';

import { API_ROUTES_PET_INTERACTIONS } from './api-routes';

export type TCarePetBody = {
  type: 'STROKE' | 'FEED';
};

export type TCarePetResponse = Pick<
  TPet,
  | 'happiness'
  | 'happinessMultiplier'
  | 'calculatedAt'
  | 'decaysToZeroAt'
  | 'strokeNextAvailableAt'
  | 'feedNextAvailableAt'
>;

export const carePetPost = (body: TCarePetBody): Promise<TCarePetResponse> => {
  return apiRequest<TCarePetResponse>(
    fetch(API_ROUTES_PET_INTERACTIONS.care, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        ...getAuthHeaders(),
      },
      body: JSON.stringify(body),
    }),
  );
};
