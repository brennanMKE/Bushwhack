// Candle flicker for lit pattern swatches. Each flame runs its own seeded
// noise, so no two boxes flicker alike; one shared animation loop drives
// every flame that is on screen and stops when none are.

export interface FlameState {
	k: number; // brightness, 0.55–1
	sway: number; // sideways drift of the flame, −1–1
}

type Draw = (s: FlameState) => void;

/** Steady state, for reduced motion and before the first frame. */
export const STILL: FlameState = { k: 0.9, sway: 0 };

function rng(seed: number) {
	let s = seed >>> 0 || 1;
	return () => {
		s = (Math.imul(s, 1664525) + 1013904223) >>> 0;
		return s / 4294967296;
	};
}

/** Smooth 1D value noise in −1…1. */
function noise(rand: () => number) {
	const pts = Float32Array.from({ length: 256 }, rand);
	return (x: number) => {
		const i = Math.floor(x);
		const f = x - i;
		const u = f * f * (3 - 2 * f);
		return (pts[i & 255] * (1 - u) + pts[(i + 1) & 255] * u) * 2 - 1;
	};
}

/** A candle: slow breathing, flutter and shimmer, plus the odd gutter. */
export function flame(seed = Math.random() * 2 ** 32) {
	const rand = rng(seed);
	const breathe = noise(rand);
	const flutter = noise(rand);
	const shimmer = noise(rand);
	const drift = noise(rand);
	const offset = rand() * 1000; // so flames don't start in step
	let nextGutter = 4 + rand() * 12;
	let gutterAt = -10;
	let gutterLen = 0.5;
	let gutterDepth = 0.2;
	return (seconds: number): FlameState => {
		const t = seconds + offset;
		if (seconds > nextGutter) {
			gutterAt = seconds;
			gutterLen = 0.25 + rand() * 0.6;
			gutterDepth = 0.12 + rand() * 0.18;
			nextGutter = seconds + 6 + rand() * 16;
		}
		const g = (seconds - gutterAt) / gutterLen;
		const gutter = g >= 0 && g <= 1 ? Math.sin(Math.PI * g) * gutterDepth : 0;
		const k = 0.84 + 0.08 * breathe(t * 0.5) + 0.05 * flutter(t * 2.7) + 0.025 * shimmer(t * 9) - gutter;
		return {
			k: Math.min(1, Math.max(0.55, k)),
			sway: Math.max(-1, Math.min(1, drift(t * 0.7) + 0.4 * flutter(t * 2.7 + 50)))
		};
	};
}

// One loop for every visible flame, at about 30 fps: plenty for a candle.
const active = new Map<Draw, (s: number) => FlameState>();
let raf = 0;
let last = 0;

function frame(now: number) {
	raf = requestAnimationFrame(frame);
	if (now - last < 33) return;
	last = now;
	for (const [draw, f] of active) draw(f(now / 1000));
}

/** Start drawing a flame; returns a function that stops it. */
export function light(draw: Draw, f = flame()): () => void {
	active.set(draw, f);
	if (!raf) raf = requestAnimationFrame(frame);
	return () => {
		active.delete(draw);
		if (!active.size) {
			cancelAnimationFrame(raf);
			raf = 0;
		}
	};
}

export const reducedMotion = () =>
	typeof matchMedia === 'function' && matchMedia('(prefers-reduced-motion: reduce)').matches;
