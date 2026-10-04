// Routers, guide bushings and bits for the workbench presets and the Gear
// page. Sizes are as published by the maker or a distributor (see each
// source); null means we couldn't confirm the number. Measure your own
// bushing with calipers before you trust any of them. No product links:
// Bushwhack doesn't sell or recommend any brand.

import { parseLength, fraction } from './units.ts';

/** Where a number came from, for whoever maintains this list. Not shown. */
export interface Source {
	label: string;
}

/** How bushings attach to a router base. */
export type System = 'pc' | 'bosch' | 'festool' | 'triton';

export const systems: Record<System, string> = {
	pc: 'Porter-Cable style: two-piece threaded bushing with a lock nut, in a 1-3/16 in hole with a 1-3/8 in counterbore',
	bosch: 'Bosch quick-change (RA1126 adapter)',
	festool: 'Festool guide bushes',
	triton: 'Triton TGA250 guide bushes'
};

export interface Router {
	id: string;
	name: string;
	kind: string;
	collets: string;
	/** Bushing systems it takes, natively or through the noted adapter. */
	takes: System[];
	note: string;
	source: Source[];
}

export interface Bushing {
	id: string;
	part: string;
	od: string; // as entered in the workbench: "5/16", "17mm"
	id_: string | null; // inside diameter: the bit must pass through it
	length: string | null; // how far it sticks out below the base
}

export interface BushingSet {
	id: string;
	name: string;
	system: System;
	note: string;
	bushings: Bushing[];
	source: Source[];
}

export interface Bit {
	id: string;
	name: string;
	dia: string;
	kind: string;
	cut: string | null;
	overall: string | null;
	source: Source[];
}

const bosch = (page: string, label = 'Bosch'): Source => ({ label: `${label}: boschtools.com ${page}` });

export const routers: Router[] = [
	{
		id: 'bosch-colt',
		name: 'Bosch Colt GKF125CE',
		kind: 'Palm',
		collets: '1/4 in',
		takes: ['bosch', 'pc'],
		note: 'The plunge base takes the RA1126 quick-change adapter, so the RA1128 set fits directly. The fixed base needs the PR110 subbase.',
		source: [bosch('us/en/products/gkf125cepk-0601628111'), bosch('us/en/products/pr110-1600A009PF', 'Bosch PR110')]
	},
	{
		id: 'bosch-1617',
		name: 'Bosch 1617EVS',
		kind: '2-1/4 hp fixed base',
		collets: '1/4, 3/8, 1/2 in, 8 mm',
		takes: ['bosch', 'pc'],
		note: 'Takes guides through the RA1129 adapter set (RA1126 plus the RA1100 Porter-Cable adapter), which the RA1128 set also includes.',
		source: [bosch('us/en/products/1617evs-0601617763'), bosch('us/en/products/ra1129-2610041329', 'Bosch RA1129')]
	},
	{
		id: 'bosch-mrc23',
		name: 'Bosch MRC23EVSK',
		kind: '2.3 hp fixed and plunge',
		collets: '1/4, 1/2 in',
		takes: ['bosch', 'pc'],
		note: 'Takes guides through the RA1129 adapter set, or the adapters in the RA1128 set.',
		source: [bosch('ca/en/products/mrc23evsk-0601624012')]
	},
	{
		id: 'dewalt-611',
		name: 'DeWalt DWP611 / DCW600',
		kind: 'Compact',
		collets: '1/4 in',
		takes: ['pc'],
		note: 'Takes Porter-Cable style guides natively. DeWalt points to the Porter-Cable 42000 kit.',
		source: [{ label: 'DeWalt support'}]
	},
	{
		id: 'dewalt-618',
		name: 'DeWalt DW618',
		kind: '2-1/4 hp',
		collets: '1/4, 1/2 in',
		takes: ['pc'],
		note: 'Its subbase takes Porter-Cable style guides.',
		source: [{ label: 'Ace Tool'}]
	},
	{
		id: 'makita-rt0701c',
		name: 'Makita RT0701C',
		kind: 'Compact',
		collets: '1/4, 3/8 in',
		takes: ['pc'],
		note: 'Takes Porter-Cable style guides through the 321492-3 template guide adapter.',
		source: [{ label: 'Hanes Supply'}]
	},
	{
		id: 'makita-xtr01',
		name: 'Makita XTR01',
		kind: 'Compact, 18 V',
		collets: '1/4, 3/8 in',
		takes: ['pc'],
		note: 'Comes with template guide 343577-5 and accepts standard template guides.',
		source: [{ label: 'Toolnut'}]
	},
	{
		id: 'milwaukee-2723',
		name: 'Milwaukee 2723-20 M18 Fuel',
		kind: 'Compact, 18 V',
		collets: '1/4 in',
		takes: ['pc'],
		note: 'Use the included template subbase, which takes guides up to 1-3/16 in. Centre it with a centring cone.',
		source: [
			{ label: 'Milwaukee'},
			{ label: 'Manual'}
		]
	},
	{
		id: 'milwaukee-5615',
		name: 'Milwaukee 5615 / 5616',
		kind: '1-3/4 hp',
		collets: '1/4, 1/2 in',
		takes: ['pc'],
		note: 'Takes Porter-Cable style guides with the 49-54-1040 subbase.',
		source: [{ label: 'Milwaukee 49-54-1040'}]
	},
	{
		id: 'pc-690',
		name: 'Porter-Cable 690LR',
		kind: '1-3/4 hp',
		collets: '1/4, 1/2 in',
		takes: ['pc'],
		note: 'The original Porter-Cable style base.',
		source: [{ label: 'Porter-Cable'}]
	},
	{
		id: 'pc-450',
		name: 'Porter-Cable 450',
		kind: 'Compact',
		collets: '1/4 in',
		takes: ['pc'],
		note: 'Its 4 in subbase takes standard template guides.',
		source: [{ label: 'Grizzly'}]
	},
	{
		id: 'festool-of1010',
		name: 'Festool OF 1010',
		kind: 'Plunge',
		collets: '8 mm, 1/4 in',
		takes: ['festool', 'pc'],
		note: 'Takes Festool guide bushes. Adapter 469625 takes Porter-Cable style guides.',
		source: [{ label: 'Festool Owners Group'}]
	},
	{
		id: 'festool-of1400',
		name: 'Festool OF 1400',
		kind: 'Plunge',
		collets: '8 mm, 1/4 in',
		takes: ['festool', 'pc'],
		note: 'Takes Festool guide bushes. Adapter 493566 takes Porter-Cable style guides.',
		source: [{ label: 'Festool Owners Group'}]
	},
	{
		id: 'triton-tra001',
		name: 'Triton TRA001',
		kind: '2-1/4 hp plunge',
		collets: '1/4, 1/2 in',
		takes: ['triton', 'pc'],
		note: "Uses the TGA250 guide bush kit. Its alignment bush also takes Porter-Cable style guides.",
		source: [{ label: 'Triton'}]
	},
	{
		id: 'ryobi-p601',
		name: 'Ryobi P601',
		kind: 'Trim, 18 V',
		collets: '1/4 in',
		takes: [],
		note: "Can't take guide bushings as shipped: its base is square with no guide hole. An aftermarket round base, such as the Woodpeckers SMTROB, adds one.",
		source: [
			{ label: 'Pro Tool Reviews'},
			{ label: 'Woodpeckers base'}
		]
	}
];

export const bushingSets: BushingSet[] = [
	{
		id: 'bosch-ra1128',
		name: 'Bosch RA1128',
		system: 'bosch',
		note: 'Includes the RA1126 quick-change adapter and the RA1100 Porter-Cable adapter. Fits Bosch 1613, 1614, 1617, 1618, 1619, PR10/20E, MR23EVS and the Colt plunge base.',
		bushings: [
			{ id: 'ra1103', part: 'RA1103', od: '5/16', id_: '17/64', length: '9/64' },
			{ id: 'ra1105', part: 'RA1105', od: '7/16', id_: '3/8', length: '9/64' },
			{ id: 'ra1109', part: 'RA1109', od: '1/2', id_: '13/32', length: '7/16' },
			{ id: 'ra1113', part: 'RA1113', od: '5/8', id_: '17/32', length: '1/2' },
			{ id: 'ra1115', part: 'RA1115', od: '3/4', id_: '21/32', length: '3/16' },
			{ id: 'ra1121', part: 'RA1121', od: '1 3/8', id_: '1 19/64', length: '7/16' }
		],
		source: [bosch('us/en/products/ra1128-2610041328'), { label: 'Toolnut (RA1121)'}]
	},
	{
		id: 'pc-42000',
		name: 'Porter-Cable 42000',
		system: 'pc',
		note: 'Seven guides and two lock nuts. The standard most other routers copy.',
		bushings: [
			{ id: 'pc42054', part: '42054', od: '5/16', id_: '1/4', length: null },
			{ id: 'pc42036', part: '42036', od: '3/8', id_: '9/32', length: '5/16' },
			{ id: 'pc42027', part: '42027', od: '7/16', id_: '11/32', length: '5/32' },
			{ id: 'pc42033', part: '42033', od: '1/2', id_: '13/32', length: null },
			{ id: 'pc42045', part: '42045', od: '5/8', id_: '17/32', length: '9/16' },
			{ id: 'pc42024', part: '42024', od: '3/4', id_: '21/32', length: '9/16' },
			{ id: 'pc42042', part: '42042', od: '51/64', id_: '5/8', length: null }
		],
		source: [
			{ label: 'Porter-Cable'},
			{ label: 'Hanes Supply'}
		]
	},
	{
		id: 'milescraft-1228',
		name: 'Milescraft 1228',
		system: 'pc',
		note: 'Seven steel bushings. Inside diameters not published.',
		bushings: [
			{ id: 'mc-516', part: '5/16', od: '5/16', id_: null, length: null },
			{ id: 'mc-38', part: '3/8', od: '3/8', id_: null, length: null },
			{ id: 'mc-716', part: '7/16', od: '7/16', id_: null, length: null },
			{ id: 'mc-12', part: '1/2', od: '1/2', id_: null, length: null },
			{ id: 'mc-58', part: '5/8', od: '5/8', id_: null, length: null },
			{ id: 'mc-34', part: '3/4', od: '3/4', id_: null, length: null },
			{ id: 'mc-5164', part: '51/64', od: '51/64', id_: null, length: null }
		],
		source: [{ label: 'Walmart'}]
	},
	{
		id: 'whiteside-9500',
		name: 'Whiteside 9500 inlay kit',
		system: 'pc',
		note: 'For one-template inlays: rout the pocket with the collar on, and the insert with it off. Comes with the RD1600 1/8 in downcut bit.',
		bushings: [
			{ id: 'ws9500', part: 'bushing', od: '5/16', id_: '3/16', length: '3/16' },
			{ id: 'ws9500-collar', part: 'with collar', od: '9/16', id_: '3/16', length: '3/16' }
		],
		source: [{ label: 'Highland Woodworking'}]
	},
	{
		id: 'triton-tga250',
		name: 'Triton TGA250',
		system: 'triton',
		note: 'For Triton routers. Inside diameters not published.',
		bushings: [
			{ id: 'tr-516', part: '5/16', od: '5/16', id_: null, length: null },
			{ id: 'tr-38', part: '3/8', od: '3/8', id_: null, length: null },
			{ id: 'tr-716', part: '7/16', od: '7/16', id_: null, length: null },
			{ id: 'tr-12', part: '1/2', od: '1/2', id_: null, length: null },
			{ id: 'tr-58', part: '5/8', od: '5/8', id_: null, length: null },
			{ id: 'tr-5164', part: '51/64', od: '51/64', id_: null, length: null },
			{ id: 'tr-34', part: '3/4', od: '3/4', id_: null, length: null },
			{ id: 'tr-30', part: '30 mm', od: '30mm', id_: null, length: null }
		],
		source: [{ label: 'Triton'}]
	},
	{
		id: 'festool-of1010',
		name: 'Festool guide bushes (OF 1010)',
		system: 'festool',
		note: 'Sizes from Festool owners; only the KR-D 17 inside diameter is confirmed.',
		bushings: [
			{ id: 'fe-108', part: '10.8 mm', od: '10.8mm', id_: null, length: null },
			{ id: 'fe-138', part: '13.8 mm', od: '13.8mm', id_: null, length: null },
			{ id: 'fe-17', part: 'KR-D 17', od: '17mm', id_: '14mm', length: null },
			{ id: 'fe-24', part: '24 mm', od: '24mm', id_: null, length: null },
			{ id: 'fe-27', part: '27 mm', od: '27mm', id_: null, length: null },
			{ id: 'fe-30', part: '30 mm', od: '30mm', id_: null, length: null },
			{ id: 'fe-40', part: '40 mm', od: '40mm', id_: null, length: null }
		],
		source: [{ label: 'Festool Owners Group'}]
	}
];

export const bits: Bit[] = [
	{ id: 'spetool-18u', name: 'SpeTool SPLS-D1/4-1/8-U-TAC', dia: '1/8', kind: 'Upcut spiral', cut: '1', overall: '3', source: [{ label: 'Amazon'}] },
	{ id: 'whiteside-ru1600', name: 'Whiteside RU1600', dia: '1/8', kind: 'Upcut spiral', cut: '1/2', overall: null, source: [{ label: "Woodworker's Source"}] },
	{ id: 'whiteside-rd1600', name: 'Whiteside RD1600', dia: '1/8', kind: 'Downcut spiral', cut: '1/2', overall: '2', source: [{ label: 'Highland Woodworking'}] },
	{ id: 'amana-46125k', name: 'Amana 46125-K', dia: '1/8', kind: 'Upcut spiral', cut: '13/16', overall: '2 1/2', source: [{ label: 'Carbatec'}] },
	{ id: 'freud-75100', name: 'Freud 75-100', dia: '1/8', kind: 'Upcut spiral', cut: '1/2', overall: '2', source: [{ label: 'Toolnut' }] },
	{ id: 'bosch-85908mc', name: 'Bosch 85908MC', dia: '1/8', kind: 'Upcut spiral', cut: '1/2', overall: null, source: [{ label: 'Automa'}] },
	{ id: 'cmt-19100111', name: 'CMT 191.001.11', dia: '1/8', kind: 'Upcut spiral', cut: '1/2', overall: '2', source: [{ label: 'Ace Tool' }] },
	{ id: 'yonico-31210sc', name: 'Yonico 31210-SC', dia: '1/8', kind: 'Upcut spiral', cut: '1/2', overall: '2', source: [{ label: 'Precision Bits'}] },
	{ id: 'whiteside-ru1800', name: 'Whiteside RU1800', dia: '3/16', kind: 'Upcut spiral', cut: '3/4', overall: '2 1/2', source: [{ label: 'EOA Saw'}] },
	{ id: 'amana-46101', name: 'Amana 46101', dia: '3/16', kind: 'Upcut spiral', cut: '3/4', overall: '2', source: [{ label: 'Burns Tools'}] },
	{ id: 'whiteside-ru2100', name: 'Whiteside RU2100', dia: '1/4', kind: 'Upcut spiral', cut: '1', overall: null, source: [{ label: 'Woodcraft' }] },
	{ id: 'freud-75102', name: 'Freud 75-102', dia: '1/4', kind: 'Upcut spiral', cut: '1', overall: '2 1/2', source: [{ label: 'Benefast' }] },
	{ id: 'bosch-85911m', name: 'Bosch 85911M', dia: '1/4', kind: 'Upcut spiral', cut: '1', overall: '2 1/2', source: [{ label: 'Beaver Tools' }] }
];

// The established inside diameter for each common outside diameter
// (Porter-Cable 42000 sizes, which most makers follow, plus Festool's
// 17 mm). Used when a bushing isn't picked from a set. Mirrors
// template.StandardBushingID.
const standard: [string, string][] = [
	['5/16', '1/4'],
	['3/8', '9/32'],
	['7/16', '11/32'],
	['1/2', '13/32'],
	['5/8', '17/32'],
	['3/4', '21/32'],
	['51/64', '5/8'],
	['1', '25/32'],
	['17mm', '14mm']
];

/** The standard inside diameter for a bushing OD, if it's a standard size. */
export function standardInner(od: string): string | null {
	return standard.find(([o]) => same(o, od))?.[1] ?? null;
}

export const inches = (s: string | null) => (s ? (parseLength(s)?.inches ?? null) : null);

/** The bushing in the given set, by id. */
export function findBushing(id: string): { set: BushingSet; bushing: Bushing } | null {
	for (const set of bushingSets) {
		const bushing = set.bushings.find((b) => b.id === id);
		if (bushing) return { set, bushing };
	}
	return null;
}

/** Sets a router can use, its native system first. */
export function setsFor(routerId: string): BushingSet[] {
	const r = routers.find((x) => x.id === routerId);
	if (!r) return bushingSets;
	return bushingSets
		.filter((s) => r.takes.includes(s.system))
		.sort((a, b) => r.takes.indexOf(a.system) - r.takes.indexOf(b.system));
}

/** Whether two lengths are the same to a thousandth of an inch. */
export const same = (a: string, b: string) => {
	const x = inches(a);
	const y = inches(b);
	return x !== null && y !== null && Math.abs(x - y) < 0.001;
};

/** The picked bushing, if it still matches the OD in the field. */
export function pickedBushing(pick: string, od: string) {
	const f = findBushing(pick);
	return f && same(f.bushing.od, od) ? f : null;
}

/** The picked bit, if it still matches the diameter in the field. */
export function pickedBit(pick: string, dia: string) {
	const b = bits.find((x) => x.id === pick);
	return b && same(b.dia, dia) ? b : null;
}

export interface Inner {
	value: string; // '' when unknown
	from: string; // where it came from, for the field's hint
}

/** The bushing's inside diameter: typed, from the picked set, or the standard. */
export function innerFor(od: string, typed: string, pick: string): Inner {
	if (typed.trim()) return { value: typed, from: '' };
	const p = pickedBushing(pick, od);
	if (p?.bushing.id_) return { value: p.bushing.id_, from: `${p.set.name} ${p.bushing.part}` };
	const std = standardInner(od);
	if (std) return { value: std, from: 'standard for this size' };
	return { value: '', from: '' };
}

export interface Fit {
	ok: boolean;
	error?: string;
	warn?: string;
}

const show = (v: number) => {
	const f = fraction(v);
	return f.includes('.') ? `${v.toFixed(3)} in` : f;
};

/** Whether the bit passes through the bushing. Mirrors template.validate. */
export function bitFit(od: string, inner: string, bit: string): Fit {
	const o = inches(od);
	const d = inches(bit);
	if (!o || !d) return { ok: false };
	if (d >= o) return { ok: false, error: `A ${show(d)} bit won't fit through a ${show(o)} bushing.` };
	const i = inches(inner);
	if (!i) return { ok: false, error: "Enter the bushing's inside diameter so Bushwhack can check the bit fits through it." };
	if (i >= o) return { ok: false, error: "The inside diameter must be smaller than the outside diameter." };
	if (d >= i)
		return { ok: false, error: `A ${show(d)} bit won't pass through this bushing's ${show(i)} inside diameter. Use a smaller bit or a bigger bushing.` };
	if (i - d < 1 / 32 - 1e-9)
		return { ok: true, warn: `Only ${show(i - d)} of clearance around the bit. Centre the bushing carefully.` };
	return { ok: true };
}
