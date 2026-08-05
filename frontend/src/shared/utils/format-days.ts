const dayForms: Record<Intl.LDMLPluralRule, string> = {
  zero: "дней",
  one: "день",
  two: "дня",
  few: "дня",
  many: "дней",
  other: "дней",
};

export const formatDays = (count: number) => {
  const rule = new Intl.PluralRules("ru-RU").select(count);

  return `${count} ${dayForms[rule]}`;
};
