<script lang="ts">
	// The workbench canvas: the result on a green cutting mat with a 1 in grid
	// and inch rulers. Hardboard template with a brass edge, dashed chalk
	// drawing, solid red cut, red hatch where the bit can't reach, and a
	// true-size brass bushing at each thin bridge. Zoom with buttons, wheel or
	// pinch; pan by dragging.
	import type { ProcessResult } from '#lib/api.ts';

	let {
		result,
		busy = false,
		highlight = null,
		show = { drawing: true, template: true, cut: true },
		pulse = { bridge: -1, key: 0 },
		label = ''
	}: {
		result: ProcessResult;
		busy?: boolean;
		highlight?: number | null;
		show?: { drawing: boolean; template: boolean; cut: boolean };
		pulse?: { bridge: number; key: number };
		label?: string;
	} = $props();

	let cw = $state(800);
	let ch = $state(560);
	let svgEl = $state<SVGSVGElement | undefined>();
	let vb = $state({ x: 0, y: 0, w: 6, h: 4 });
	let fittedFor = '';

	const RULER = 24; // px

	const pageW = $derived(result.page.widthIn);
	const pageH = $derived(result.page.heightIn);

	function fit() {
		const m = Math.max(pageW, pageH) * 0.06 + 0.25;
		// leave room for the rulers
		const px = Math.max((pageW + 2 * m) / Math.max(1, cw - RULER), (pageH + 2 * m) / Math.max(1, ch - RULER));
		const padX = RULER * px;
		vb = { x: -m - padX, y: -m - padX, w: pageW + 2 * m + padX, h: pageH + 2 * m + padX };
	}

	// Refit only when the page itself changes, so tweaking options keeps the view.
	$effect(() => {
		const key = `${pageW.toFixed(4)}x${pageH.toFixed(4)}`;
		if (key !== fittedFor && cw > 0) {
			fittedFor = key;
			fit();
		}
	});

	// inches per screen pixel, and the full visible area (meet scaling).
	const px = $derived(Math.max(vb.w / Math.max(1, cw), vb.h / Math.max(1, ch)));
	const vis = $derived({
		x: vb.x - (cw * px - vb.w) / 2,
		y: vb.y - (ch * px - vb.h) / 2,
		w: cw * px,
		h: ch * px
	});

	const inchPx = $derived(1 / px);
	const labelEvery = $derived(inchPx >= 28 ? 1 : inchPx >= 12 ? 5 : inchPx >= 3 ? 10 : 50);
	const minor = $derived(inchPx >= 64 ? 0.25 : inchPx >= 28 ? 0.5 : 0);
	const range = (a: number, b: number, step: number) => {
		const out: number[] = [];
		if (step <= 0) return out;
		for (let v = Math.ceil(a / step) * step; v <= b && out.length < 2000; v += step) out.push(+v.toFixed(4));
		return out;
	};
	const gridX = $derived(range(vis.x, vis.x + vis.w, labelEvery > 1 ? labelEvery : 1));
	const gridY = $derived(range(vis.y, vis.y + vis.h, labelEvery > 1 ? labelEvery : 1));
	const ticksX = $derived(range(vis.x, vis.x + vis.w, minor || labelEvery));
	const ticksY = $derived(range(vis.y, vis.y + vis.h, minor || labelEvery));
	const isMajor = (v: number) => Math.abs(v / labelEvery - Math.round(v / labelEvery)) < 1e-6;

	const sheet = $derived(`M0 0H${pageW}V${pageH}H0Z${result.layers.template}`);
	const bushR = $derived(result.options.bushingOD / 2);
	const thin = $derived(
		result.bridges
			.map((b, i) => ({ ...b, i }))
			.filter((b) => b.distIn > 0 && b.distIn < result.options.minBridge)
	);
	const hl = $derived(highlight == null ? null : result.shapes.find((s) => s.index === highlight));

	function zoomAt(factor: number, cx: number, cy: number) {
		const w = Math.min(Math.max(vb.w * factor, 0.2), Math.max(pageW, pageH) * 8 + 4);
		const k = w / vb.w;
		vb = { x: cx - (cx - vb.x) * k, y: cy - (cy - vb.y) * k, w, h: vb.h * k };
	}
	const zoomCenter = (factor: number) => zoomAt(factor, vb.x + vb.w / 2, vb.y + vb.h / 2);

	function toUser(clientX: number, clientY: number) {
		const r = svgEl!.getBoundingClientRect();
		return { x: vis.x + (clientX - r.left) * px, y: vis.y + (clientY - r.top) * px };
	}

	$effect(() => {
		const el = svgEl;
		if (!el) return;
		const onWheel = (e: WheelEvent) => {
			e.preventDefault();
			const p = toUser(e.clientX, e.clientY);
			zoomAt(Math.exp(e.deltaY * (e.ctrlKey ? 0.01 : 0.0015)), p.x, p.y);
		};
		el.addEventListener('wheel', onWheel, { passive: false });
		return () => el.removeEventListener('wheel', onWheel);
	});

	const pointers = new Map<number, { x: number; y: number }>();
	let dragging = $state(false);

	function down(e: PointerEvent) {
		svgEl!.setPointerCapture(e.pointerId);
		pointers.set(e.pointerId, { x: e.clientX, y: e.clientY });
		dragging = true;
	}
	function move(e: PointerEvent) {
		const prev = pointers.get(e.pointerId);
		if (!prev) return;
		if (pointers.size === 1) {
			vb = { ...vb, x: vb.x - (e.clientX - prev.x) * px, y: vb.y - (e.clientY - prev.y) * px };
		} else if (pointers.size === 2) {
			const [a, b] = [...pointers.entries()];
			const other = a[0] === e.pointerId ? b[1] : a[1];
			const d0 = Math.hypot(prev.x - other.x, prev.y - other.y);
			const d1 = Math.hypot(e.clientX - other.x, e.clientY - other.y);
			if (d0 > 0 && d1 > 0) {
				const mid = toUser((e.clientX + other.x) / 2, (e.clientY + other.y) / 2);
				zoomAt(d0 / d1, mid.x, mid.y);
			}
		}
		pointers.set(e.pointerId, { x: e.clientX, y: e.clientY });
	}
	function up(e: PointerEvent) {
		pointers.delete(e.pointerId);
		if (pointers.size === 0) dragging = false;
	}
	function key(e: KeyboardEvent) {
		const step = vb.w * 0.1;
		const moves: Record<string, () => void> = {
			'+': () => zoomCenter(0.8),
			'=': () => zoomCenter(0.8),
			'-': () => zoomCenter(1.25),
			'0': fit,
			ArrowLeft: () => (vb = { ...vb, x: vb.x - step }),
			ArrowRight: () => (vb = { ...vb, x: vb.x + step }),
			ArrowUp: () => (vb = { ...vb, y: vb.y - step }),
			ArrowDown: () => (vb = { ...vb, y: vb.y + step })
		};
		if (moves[e.key]) {
			e.preventDefault();
			moves[e.key]();
		}
	}
	const fmt = (v: number) => (Number.isInteger(v) ? String(v) : v.toFixed(2).replace(/0$/, ''));
</script>

<div class="mat" class:busy bind:clientWidth={cw} bind:clientHeight={ch}>
	<!-- svelte-ignore a11y_no_noninteractive_tabindex, a11y_no_noninteractive_element_interactions -->
	<svg
		bind:this={svgEl}
		viewBox="{vb.x} {vb.y} {vb.w} {vb.h}"
		class:dragging
		role="img"
		aria-label={label}
		tabindex="0"
		onpointerdown={down}
		onpointermove={move}
		onpointerup={up}
		onpointercancel={up}
		onkeydown={key}
	>
		<defs>
			<pattern id="lost-hatch" patternUnits="userSpaceOnUse" width={5 * px} height={5 * px} patternTransform="rotate(45)">
				<line x1="0" y1="0" x2="0" y2={5 * px} stroke="var(--cut-on-mat)" stroke-width={1.4 * px} />
			</pattern>
		</defs>
		<rect x={vis.x} y={vis.y} width={vis.w} height={vis.h} fill="var(--mat)" />
		<g stroke="var(--mat-grid)" stroke-width={px} opacity="0.75">
			{#each gridX as gx (gx)}<line x1={gx} y1={vis.y} x2={gx} y2={vis.y + vis.h} />{/each}
			{#each gridY as gy (gy)}<line x1={vis.x} y1={gy} x2={vis.x + vis.w} y2={gy} />{/each}
		</g>

		{#if show.template}
			<path d={sheet} fill="var(--hardboard)" fill-rule="evenodd" />
			<path d={result.layers.template} fill="none" stroke="var(--brass)" stroke-width={px} />
		{/if}
		{#if result.layers.lost}
			<path d={result.layers.lost} fill="url(#lost-hatch)" stroke="var(--cut-on-mat)" stroke-width={0.75 * px} />
		{/if}
		{#if hl}
			<path d={hl.path} fill="var(--brass)" fill-opacity="0.28" stroke="var(--brass)" stroke-width={3 * px} />
		{/if}
		{#if show.cut}
			<path d={result.layers.cut} fill="none" stroke="var(--cut-on-mat)" stroke-width={2 * px} stroke-linejoin="round" />
		{/if}
		{#if show.drawing}
			<path d={result.layers.drawing} fill="none" stroke="var(--chalk-on-mat)" stroke-width={1.8 * px} stroke-dasharray="{6 * px} {4 * px}" stroke-linejoin="round" />
		{/if}

		{#each thin as b (b.i)}
			{@const lx = b.at.xIn + bushR + 26 * px}
			{@const ly = b.at.yIn - bushR - 26 * px}
			{#key pulse.bridge === b.i ? pulse.key : 0}
				<g class="bridge" class:pulse={pulse.bridge === b.i && pulse.key > 0}>
					<circle cx={b.at.xIn} cy={b.at.yIn} r={bushR} fill="var(--brass)" fill-opacity="0.18" stroke="var(--brass)" stroke-width={2 * px} />
					<line x1={b.at.xIn + bushR * 0.71} y1={b.at.yIn - bushR * 0.71} x2={lx} y2={ly} stroke="var(--brass)" stroke-width={1.5 * px} />
					<rect x={lx} y={ly - 11 * px} width={58 * px} height={18 * px} fill="var(--mat)" stroke="var(--brass)" stroke-width={px} />
					<text x={lx + 5 * px} y={ly + 2 * px} font-size={12 * px} fill="var(--shop-wall)" font-weight="700">{b.distIn.toFixed(3)} in</text>
				</g>
			{/key}
		{/each}

		<!-- rulers, drawn in screen-sized units along the visible edges -->
		<g class="rulers" font-size={10.5 * px} fill="var(--shop-wall)">
			<rect x={vis.x} y={vis.y} width={vis.w} height={RULER * px} fill="var(--mat)" opacity="0.94" />
			<rect x={vis.x} y={vis.y} width={RULER * px} height={vis.h} fill="var(--mat)" opacity="0.94" />
			<line x1={vis.x} y1={vis.y + RULER * px} x2={vis.x + vis.w} y2={vis.y + RULER * px} stroke="var(--mat-grid)" stroke-width={px} />
			<line x1={vis.x + RULER * px} y1={vis.y} x2={vis.x + RULER * px} y2={vis.y + vis.h} stroke="var(--mat-grid)" stroke-width={px} />
			{#each ticksX as tx (tx)}
				{#if tx > vis.x + RULER * px}
					<line x1={tx} y1={vis.y + RULER * px} x2={tx} y2={vis.y + RULER * px * (isMajor(tx) ? 0.45 : 0.72)} stroke="var(--shop-wall)" stroke-width={px} />
					{#if isMajor(tx)}<text x={tx + 3 * px} y={vis.y + 11 * px}>{fmt(tx)}</text>{/if}
				{/if}
			{/each}
			{#each ticksY as ty (ty)}
				{#if ty > vis.y + RULER * px}
					<line x1={vis.x + RULER * px} y1={ty} x2={vis.x + RULER * px * (isMajor(ty) ? 0.45 : 0.72)} y2={ty} stroke="var(--shop-wall)" stroke-width={px} />
					{#if isMajor(ty)}<text x={vis.x + 3 * px} y={ty + 11 * px}>{fmt(ty)}</text>{/if}
				{/if}
			{/each}
			<rect x={vis.x} y={vis.y} width={RULER * px} height={RULER * px} fill="var(--mat)" />
			<text x={vis.x + 5 * px} y={vis.y + 15 * px} font-weight="700">in</text>
		</g>
	</svg>

	<div class="zoom" role="group" aria-label="Zoom">
		<button type="button" onclick={() => zoomCenter(0.8)} aria-label="Zoom in">+</button>
		<button type="button" onclick={() => zoomCenter(1.25)} aria-label="Zoom out">−</button>
		<button type="button" onclick={fit} aria-label="Fit to page">Fit</button>
	</div>
</div>

<style>
	.mat {
		position: relative;
		width: 100%;
		height: 100%;
		min-height: 320px;
		background: var(--mat);
		transition: opacity 150ms;
		overflow: hidden;
	}
	.mat.busy {
		opacity: 0.7;
	}
	svg {
		position: absolute;
		inset: 0;
		width: 100%;
		height: 100%;
		touch-action: none;
		cursor: grab;
		font-family: var(--font-text);
	}
	svg.dragging {
		cursor: grabbing;
	}
	svg:focus-visible {
		outline: 2px solid var(--brass);
		outline-offset: -4px;
	}
	.zoom {
		position: absolute;
		right: 0.75rem;
		bottom: 0.75rem;
		display: flex;
		gap: 4px;
	}
	.zoom button {
		min-width: 2.5rem;
		height: 2.5rem;
		border-radius: var(--radius);
		border: 1.5px solid var(--brass);
		background: var(--mat);
		color: var(--shop-wall);
		font: inherit;
		font-weight: 700;
		font-size: 1.1rem;
		cursor: pointer;
	}
	.zoom button:hover {
		background: #285a4d;
	}
	.bridge.pulse {
		animation: pulse 900ms ease-out 1;
		transform-box: fill-box;
		transform-origin: center;
	}
	@keyframes pulse {
		0% {
			opacity: 1;
		}
		30% {
			opacity: 0.25;
		}
		60% {
			opacity: 1;
		}
		80% {
			opacity: 0.45;
		}
		100% {
			opacity: 1;
		}
	}
</style>
