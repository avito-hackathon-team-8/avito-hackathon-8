import { describe, expect, it } from 'vitest';

import { formatWord } from './format-word';

describe('formatWord', () => {
  const forms = ['награда', 'награды', 'наград'] as const;

  it.each([
    [1, 'награда'],
    [2, 'награды'],
    [5, 'наград'],
    [11, 'наград'],
    [21, 'награда'],
    [22, 'награды'],
    [25, 'наград'],
  ])('выбирает форму слова для числа %i', (count, expected) => {
    expect(formatWord(count, forms)).toBe(expected);
  });
});
