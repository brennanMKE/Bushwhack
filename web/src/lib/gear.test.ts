import { describe, expect, it } from 'vitest';
import { bitFit, bushingSets, innerFor, inches, routers, setsFor, standardInner } from './gear.ts';

describe('gear', () => {
	it('every listed size parses, and each bit passes its own bushing', () => {
		for (const s of bushingSets) {
			for (const b of s.bushings) {
				expect(inches(b.od), `${s.name} ${b.part}`).toBeGreaterThan(0);
				if (b.id_) expect(inches(b.id_)!).toBeLessThan(inches(b.od)!);
			}
		}
	});
	it('uses the standard inside diameter for a typed size', () => {
		expect(standardInner('0.3125')).toBe('1/4');
		expect(innerFor('5/16', '', '').value).toBe('1/4');
		expect(innerFor('5/16', '', 'ra1103').value).toBe('17/64');
		expect(innerFor('5/16', '0.27', 'ra1103').value).toBe('0.27');
		expect(innerFor('0.55', '', '').value).toBe('');
	});
	it('rejects a bit that will not pass through', () => {
		expect(bitFit('5/16', '1/4', '1/8').ok).toBe(true);
		expect(bitFit('5/16', '1/4', '1/4').ok).toBe(false);
		expect(bitFit('5/16', '17/64', '1/4')).toMatchObject({ ok: true, warn: expect.stringContaining('1/64 in') });
		expect(bitFit('0.55', '', '1/8').ok).toBe(false);
		expect(bitFit('1/2', '5/8', '1/8').ok).toBe(false);
	});
	it('offers only sets a router can take', () => {
		expect(setsFor('dewalt-611').every((s) => s.system === 'pc')).toBe(true);
		expect(setsFor('bosch-1617')[0].system).toBe('bosch');
		expect(setsFor('ryobi-p601')).toEqual([]);
		expect(routers.length).toBeGreaterThan(10);
	});
});
