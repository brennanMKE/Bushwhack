// Whether pattern swatches show as lit boxes. Remembered per browser.

const KEY = 'bushwhack.lights.v1';

// Starts on, matching the prerendered HTML; restoreLights() applies the
// saved choice after hydration.
export const lights = $state({ on: true });

let restored = false;

export function restoreLights() {
	if (restored) return;
	restored = true;
	try {
		lights.on = localStorage.getItem(KEY) !== 'off';
	} catch {
		// storage blocked: stay on
	}
}

export function setLights(on: boolean) {
	restored = true;
	lights.on = on;
	try {
		localStorage.setItem(KEY, on ? 'on' : 'off');
	} catch {
		// private mode: lasts for this visit only
	}
}
