import { describe, expect, it } from 'vitest';
import { combos, holeOffset, pieceOffset } from './offsets.ts';
import { fraction } from './units.ts';

describe('offsets', () => {
	it('matches the worked example', () => {
		expect(fraction(holeOffset(5 / 16, 1 / 8))).toBe('3/32 in');
		expect(fraction(pieceOffset(5 / 16, 1 / 8))).toBe('7/32 in');
	});
	it('piece offset exceeds hole offset by the bit diameter', () => {
		for (const c of combos) {
			expect(pieceOffset(c.bushingIn, c.bitIn) - holeOffset(c.bushingIn, c.bitIn)).toBeCloseTo(c.bitIn, 9);
		}
	});
});
