import { tasksMock } from '../model/mock/tasks-mock';

export type TTaskType =
  'OPEN_NOTIFICATIONS' | 'ADD_TO_FAVORITES' | 'PUBLISH_LISTING' | 'COMPLETE_DEAL';

export type TaskStatus = 'CLAIMED' | 'COMPLETED' | 'LOCKED';

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
