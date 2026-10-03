// Types and calls for the Go API (internal/server).

export interface Point {
	xIn: number;
	yIn: number;
}

export interface Shape {
	index: number;
	element: string;
	name: string;
	path: string;
	widthIn: number;
	heightIn: number;
	templateWidthIn: number;
	templateHeightIn: number;
	lostAreaPct: number;
	overcutPct: number;
}

export interface Bridge {
	a: number;
	b: number;
	distIn: number;
	at: Point;
	p: Point;
	q: Point;
}

export interface Note {
	kind: 'merged' | 'bridge' | 'island' | 'lost' | 'uncuttable' | 'overcut' | 'page' | 'input';
	text: string;
	shapes?: number[];
	bridge: number; // index into bridges, or -1
}

export interface Layers {
	drawing: string;
	template: string;
	cut: string;
	center: string;
	lost: string;
}

export interface ProcessResult {
	templateSvg: string;
	previewSvg: string;
	page: { widthIn: number; heightIn: number };
	scaleInPerUnit: number;
	offsetIn: number;
	shapes: Shape[];
	minBridgeIn: number | null;
	minBridgeAt: Point | null;
	bridges: Bridge[];
	warnings: string[];
	notes: Note[];
	layers: Layers;
	pattern?: string;
	options: {
		bushingOD: number;
		bitDia: number;
		mode: 'hole' | 'piece';
		fit: string;
		size: number;
		scale?: number;
		minBridge: number;
	};
}

export interface Pattern {
	slug: string;
	name: string;
	occasion: string;
	recommendedSizeIn: number;
	fit: string;
	note: string;
	license: string;
	drawing: string;
	widthIn: number;
	heightIn: number;
	minSafeSizeIn: number | null;
}

export interface Occasion {
	slug: string;
	name: string;
	description: string;
	upcoming: boolean;
	patterns: Pattern[];
}

export class ApiError extends Error {
	constructor(
		message: string,
		public status: number
	) {
		super(message);
	}
}

export async function processSvg(form: FormData, signal?: AbortSignal): Promise<ProcessResult> {
	const res = await fetch('/api/process', { method: 'POST', body: form, signal });
	let body: unknown;
	try {
		body = await res.json();
	} catch {
		throw new ApiError('The server sent something unexpected. Try again.', res.status);
	}
	if (!res.ok) {
		const msg = (body as { error?: string })?.error ?? 'Something went wrong.';
		throw new ApiError(msg, res.status);
	}
	return body as ProcessResult;
}

let patternsPromise: Promise<Occasion[]> | null = null;

/** Pattern library, ordered with the upcoming occasion first. Cached. */
export function loadPatterns(): Promise<Occasion[]> {
	patternsPromise ??= fetch('/api/patterns')
		.then((r) => {
			if (!r.ok) throw new ApiError('Could not load patterns.', r.status);
			return r.json();
		})
		.then((b: { occasions: Occasion[] }) => b.occasions)
		.catch((e) => {
			patternsPromise = null;
			throw e;
		});
	return patternsPromise;
}

export function shapeLabel(s: Shape): string {
	const n = s.name.replace(/[-_]+/g, ' ').trim();
	if (!n || /^(path|rect|circle|ellipse|polygon|layer)\b/i.test(n)) return `Shape ${s.index}`;
	return n.charAt(0).toUpperCase() + n.slice(1);
}
