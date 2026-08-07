import { tasksMock } from '../model/mock/tasks-mock';

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

export const getTasks = async (): Promise<TTask[]> => {
  return await Promise.resolve(tasksMock);
};
