export const tasksQueryKeys = {
  all: ['rewards'] as const,

  list: () => [...tasksQueryKeys.all, 'list'] as const,

  receive: () => [...tasksQueryKeys.all, 'receive'] as const,
};
