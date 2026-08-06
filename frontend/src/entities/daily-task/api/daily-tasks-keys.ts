export const tasksQueryKeys = {
  all: ['tasks'] as const,

  list: () => [...tasksQueryKeys.all, 'list'] as const,

  progress: () => [...tasksQueryKeys.all, 'progress'] as const,
};

export const tasksMutationKeys = {
  all: ['tasks'] as const,

  record: () => [...tasksMutationKeys.all, 'record'] as const,

  claim: (taskId: string) => [...tasksMutationKeys.all, 'claim', taskId] as const,
};
