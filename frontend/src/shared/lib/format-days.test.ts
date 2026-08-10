import { describe, expect, it } from 'vitest';

import { formatDays } from './format-days';

describe('formatDays', () => {
  it.each([
    [0, '0 дней'],
    [1, '1 день'],
    [2, '2 дня'],
    [5, '5 дней'],
    [11, '11 дней'],
    [21, '21 день'],
    [22, '22 дня'],
    [25, '25 дней'],
  ])('форматирует %i с правильным склонением', (count, expected) => {
    expect(formatDays(count)).toBe(expected);
  });
});
