// Remembers the bushing, bit, mode, minimum bridge and picked gear per
// browser. Never the uploaded files.

const KEY = 'bushwhack.settings.v1';

export interface Saved {
	bushing: string;
	bit: string;
	mode: 'hole' | 'piece';
	minBridge: string;
	inner: string; // bushing inside diameter typed by hand; '' = from the pick or the standard
	router: string; // gear.ts ids; '' = none picked
	bushingPick: string;
	bitPick: string;
}

export const defaults: Saved = {
	bushing: '5/16',
	bit: '1/8',
	mode: 'hole',
	minBridge: '1/4',
	inner: '',
	router: '',
	bushingPick: '',
	bitPick: ''
};

const str = (v: unknown, d: string) => (typeof v === 'string' ? v : d);

export function loadSettings(): Saved {
	try {
		const raw = localStorage.getItem(KEY);
		if (!raw) return { ...defaults };
		const v = JSON.parse(raw);
		return {
			bushing: str(v.bushing, defaults.bushing),
			bit: str(v.bit, defaults.bit),
			mode: v.mode === 'piece' ? 'piece' : 'hole',
			minBridge: str(v.minBridge, defaults.minBridge),
			inner: str(v.inner, ''),
			router: str(v.router, ''),
			bushingPick: str(v.bushingPick, ''),
			bitPick: str(v.bitPick, '')
		};
	} catch {
		return { ...defaults };
	}
}

export function saveSettings(s: Saved) {
	try {
		localStorage.setItem(KEY, JSON.stringify(s));
	} catch {
		// private mode or storage blocked: settings just won't stick
	}
}
