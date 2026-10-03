// Mirrors internal/units/units.go. Shared cases live in testdata/units.json.

const suffixes: [string, number][] = [
	['mm', 1 / 25.4],
	['cm', 1 / 2.54],
	['inches', 1],
	['inch', 1],
	['in', 1],
	['"', 1],
	['″', 1]
];

const numberRE = /^[+-]?(\d+\.?\d*|\.\d+)([eE][+-]?\d+)?$/;

function parseNum(s: string): number {
	s = s.trim();
	if (!numberRE.test(s)) return NaN;
	return Number(s);
}

function parseMixed(s: string): number {
	if (!s.includes('/')) return parseNum(s);
	let whole = 0;
	let frac = s;
	const i = Math.max(s.lastIndexOf(' '), s.lastIndexOf('-'));
	if (i > 0) {
		whole = parseNum(s.slice(0, i));
		if (Number.isNaN(whole)) return NaN;
		frac = s.slice(i + 1).trim();
	}
	const parts = frac.split('/');
	if (parts.length !== 2) return NaN;
	const n = parseNum(parts[0]);
	const d = parseNum(parts[1]);
	if (Number.isNaN(n) || Number.isNaN(d) || d === 0) return NaN;
	return whole < 0 ? whole - n / d : whole + n / d;
}

export interface Parsed {
	inches: number;
	unit: 'in' | 'mm' | 'cm';
	value: number; // in the entered unit
}

/** Parse "5/16", "1 1/4", "0.3125", "8mm"… Returns null if invalid. */
export function parseLength(input: string): Parsed | null {
	let s = input.trim().toLowerCase();
	let factor = 1;
	let unit: Parsed['unit'] = 'in';
	for (const [suffix, f] of suffixes) {
		if (s.endsWith(suffix)) {
			s = s.slice(0, -suffix.length).trim();
			factor = f;
			unit = suffix === 'mm' ? 'mm' : suffix === 'cm' ? 'cm' : 'in';
			break;
		}
	}
	if (s === '') return null;
	const v = parseMixed(s);
	if (!Number.isFinite(v)) return null;
	return { inches: v * factor, unit, value: v };
}

/** 0.3125 → "5/16 in"; 0.1 → "0.100 in". Mirrors units.Fraction. */
export function fraction(inches: number): string {
	const n = Math.round(inches * 64);
	if (Math.abs(n / 64 - inches) > 0.0005 || n === 0) return `${inches.toFixed(3)} in`;
	const whole = Math.trunc(n / 64);
	let num = n % 64;
	let den = 64;
	while (num !== 0 && num % 2 === 0) {
		num /= 2;
		den /= 2;
	}
	if (num === 0) return `${whole} in`;
	if (whole === 0) return `${num}/${den} in`;
	return `${whole} ${num}/${den} in`;
}

export function decimal(inches: number, places = 4): string {
	return `${inches.toFixed(places)} in`;
}

/** The echo under a measurement field: "5/16 in, 0.3125 in" or "8 mm, 0.315 in". */
export function describe(p: Parsed): string {
	if (p.unit === 'mm' || p.unit === 'cm') {
		return `${+p.value.toFixed(3)} ${p.unit}, ${p.inches.toFixed(3)} in`;
	}
	const f = fraction(p.inches);
	const d = decimal(p.inches);
	return f.includes('.') ? d : `${f}, ${d}`;
}
