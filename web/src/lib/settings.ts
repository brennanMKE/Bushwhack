// Remembers bushing, bit and mode per browser. Never the uploaded files.

const KEY = 'bushwhack.settings.v1';

export interface Saved {
	bushing: string;
	bit: string;
	mode: 'hole' | 'piece';
}

export const defaults: Saved = { bushing: '5/16', bit: '1/8', mode: 'hole' };

export function loadSettings(): Saved {
	try {
		const raw = localStorage.getItem(KEY);
		if (!raw) return { ...defaults };
		const v = JSON.parse(raw);
		return {
			bushing: typeof v.bushing === 'string' ? v.bushing : defaults.bushing,
			bit: typeof v.bit === 'string' ? v.bit : defaults.bit,
			mode: v.mode === 'piece' ? 'piece' : 'hole'
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
