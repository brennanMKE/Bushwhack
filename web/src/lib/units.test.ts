import { describe as group, expect, it } from 'vitest';
import { readFileSync } from 'node:fs';
import { fileURLToPath } from 'node:url';
import { parseLength, fraction, describe } from './units';

const cases: { input: string; inches?: number; error?: boolean }[] = JSON.parse(
	readFileSync(fileURLToPath(new URL('../../../testdata/units.json', import.meta.url)), 'utf8')
);

group('parseLength matches the Go parser', () => {
	for (const c of cases) {
		it(JSON.stringify(c.input), () => {
			const got = parseLength(c.input);
			if (c.error) expect(got).toBeNull();
			else expect(got?.inches).toBeCloseTo(c.inches!, 9);
		});
	}
});

it('formats fractions and echoes', () => {
	expect(fraction(0.3125)).toBe('5/16 in');
	expect(fraction(1.25)).toBe('1 1/4 in');
	expect(describe(parseLength('5/16')!)).toBe('5/16 in, 0.3125 in');
	expect(describe(parseLength('8mm')!)).toBe('8 mm, 0.315 in');
});
