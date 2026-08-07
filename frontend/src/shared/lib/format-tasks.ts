const dayForms: Record<Intl.LDMLPluralRule, string> = {
  zero: 'заданий',
  one: 'задание',
  two: 'задания',
  few: 'задания',
  many: 'заданий',
  other: 'заданий',
};

export const formatTasks = (count: number) => {
  const rule = new Intl.PluralRules('ru-RU').select(count);

  return dayForms[rule];
};
