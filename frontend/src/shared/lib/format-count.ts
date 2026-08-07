type TWordForms = readonly [one: string, few: string, many: string];

export const formatWord = (count: number, [one, few, many]: TWordForms) => {
  const rule = new Intl.PluralRules('ru-RU').select(count);

  const forms: Record<Intl.LDMLPluralRule, string> = {
    zero: many,
    one,
    two: few,
    few,
    many,
    other: many,
  };

  return forms[rule];
};
