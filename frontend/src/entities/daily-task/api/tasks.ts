import { apiRequest, getAuthHeaders } from '@/shared/api';
import { API_URL } from '@/shared/config';

import { API_ROUTE_DAILY_TASKS } from './api-routes';

export type TTaskType =
  | 'OPEN_NOTIFICATIONS'
  | 'VIEW_LISTINGS'
  | 'ADD_TO_FAVORITES'
  | 'PUBLISH_LISTING'
  | 'BOOST_LISTING'
  | 'LEAVE_REVIEW'
  | 'COMPLETE_DEAL'
  | 'ORDER_WITH_DELIVERY';

export type TaskStatus = 'CLAIMED' | 'COMPLETED' | 'LOCKED' | 'IN_PROGRESS';

export type TTask = {
  taskId: string;
  slot: number;
  type: TTaskType;
  description: string;
  currentCount: number;
  targetCount: number;
  rewardLeaves: number;
  requiredLevel: number;
  status: TaskStatus;
};

type TResponseGetTasks = {
  tasks: TTask[];
};

export const getTasks = async (): Promise<TTask[]> => {
  const data = await apiRequest<TResponseGetTasks>(
    fetch(`${API_URL}${API_ROUTE_DAILY_TASKS.tasks}`, {
      headers: getAuthHeaders(),
    }),
  );

  return data.tasks;
};
export const receiveTaskReward = async (id: string): Promise<TResponseGetTasks> => {
  return await apiRequest<TResponseGetTasks>(
    fetch(`${API_URL}${API_ROUTE_DAILY_TASKS.receiveRewardTask(id)}`, {
      headers: getAuthHeaders(),
      method: 'POST',
    }),
  );
};
