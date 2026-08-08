export const API_ROUTE_DAILY_TASKS = {
  tasks: '/v1/tasks',
  progress: '/v1/tasks/progress',

  receiveRewardTask: (taskId: string) => `/v1/tasks/${taskId}/claim`,
};
