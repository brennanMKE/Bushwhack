<script lang="ts">
	// The one orchestrated motion: chalk line draws on, the hardboard opening
	// grows out to the template edge, then the bushing travels the edge with
	// the red cut trailing the bit onto the chalk line. Data comes from
	// cmd/gen-demo, which runs the real pipeline.
	import { onMount } from 'svelte';
	import demo from '#lib/demo.json';

	type P = [number, number];
	const drawing = demo.drawing as P[];
	const template = demo.template as P[];
	const center = demo.center as P[];
	const cut = demo.cut as P[];
	const [vx, vy, vw, vh] = demo.viewBox;
	const bushR = demo.bushingIn / 2;
	const bitR = demo.bitIn / 2;
	const N = drawing.length;

	const T_DRAW = 1200;
	const T_GROW = 800;
	const T_TRAVEL = 3000;
	const END = T_DRAW + T_GROW + T_TRAVEL;

	let t = $state(0);
	let raf = 0;

	const clamp = (v: number) => Math.min(1, Math.max(0, v));
	const ease = (p: number) => (p < 0.5 ? 2 * p * p : 1 - Math.pow(-2 * p + 2, 2) / 2);
	const path = (pts: P[], close = true) =>
		'M' + pts.map(([x, y]) => `${x.toFixed(3)} ${y.toFixed(3)}`).join('L') + (close ? 'Z' : '');

	const chalkLen = drawing.reduce((acc, p, i) => {
		const q = drawing[(i + 1) % N];
		return acc + Math.hypot(q[0] - p[0], q[1] - p[1]);
	}, 0);

	const pDraw = $derived(clamp(t / T_DRAW));
	const pGrow = $derived(ease(clamp((t - T_DRAW) / T_GROW)));
	const pTravel = $derived(clamp((t - T_DRAW - T_GROW) / T_TRAVEL));
	const opening = $derived(
		drawing.map(([x, y], i) => [x + (template[i][0] - x) * pGrow, y + (template[i][1] - y) * pGrow] as P)
	);
	const idx = $derived(Math.min(N - 1, Math.floor(pTravel * N)));
	const bushing = $derived(center[idx]);
	const cutSoFar = $derived(pTravel >= 1 ? path(cut) : path(cut.slice(0, idx + 1), false));
	const inset = 0.28;
	const sheet = `M${vx + inset} ${vy + inset}h${vw - 2 * inset}v${vh - 2 * inset}h${-(vw - 2 * inset)}Z`;
	const done = $derived(t >= END);

	const gridX = Array.from({ length: Math.ceil(vw) + 1 }, (_, i) => Math.floor(vx) + i);
	const gridY = Array.from({ length: Math.ceil(vh) + 1 }, (_, i) => Math.floor(vy) + i);

	function play() {
		cancelAnimationFrame(raf);
		const reduce = matchMedia('(prefers-reduced-motion: reduce)').matches;
		if (reduce) {
			t = END;
			return;
		}
		const start = performance.now();
		const tick = (now: number) => {
			t = now - start;
			if (t < END) raf = requestAnimationFrame(tick);
			else t = END;
		};
		t = 0;
		raf = requestAnimationFrame(tick);
	}

	onMount(() => {
		play();
		return () => cancelAnimationFrame(raf);
	});
</script>

<figure class="demo">
	<svg viewBox="{vx} {vy} {vw} {vh}" role="img" aria-label="A drawing of a jack-o'-lantern mouth. The template opening grows past it, and a guide bushing traces the template while the router cut lands on the original line.">
		<rect x={vx} y={vy} width={vw} height={vh} fill="var(--mat)" />
		<g stroke="var(--mat-grid)" stroke-width="0.012" opacity="0.7">
			{#each gridX as gx (gx)}<line x1={gx} y1={vy} x2={gx} y2={vy + vh} />{/each}
			{#each gridY as gy (gy)}<line x1={vx} y1={gy} x2={vx + vw} y2={gy} />{/each}
		</g>
		{#if pGrow > 0}
			<path d="{sheet}{path(opening)}" fill="var(--hardboard)" fill-rule="evenodd" opacity={Math.min(1, pGrow * 1.6)} />
			<path d={path(opening)} fill="none" stroke="var(--brass)" stroke-width="0.02" opacity={pGrow} />
		{/if}
		<path
			d={path(drawing)}
			fill="none"
			stroke="var(--chalk-on-mat)"
			stroke-width="0.045"
			stroke-linejoin="round"
			stroke-dasharray={pDraw < 1 ? `${chalkLen} ${chalkLen}` : '0.16 0.09'}
			stroke-dashoffset={pDraw < 1 ? chalkLen * (1 - pDraw) : 0}
		/>
		{#if pTravel > 0}
			<path d={cutSoFar} fill="none" stroke="var(--cut-on-mat)" stroke-width="0.05" stroke-linejoin="round" stroke-linecap="round" />
		{/if}
		{#if pTravel > 0 && !done}
			<circle cx={bushing[0]} cy={bushing[1]} r={bushR} fill="var(--brass)" fill-opacity="0.25" stroke="var(--brass)" stroke-width="0.035" />
			<circle cx={bushing[0]} cy={bushing[1]} r={bitR} fill="var(--graphite)" stroke="var(--shop-wall)" stroke-width="0.012" />
		{/if}
	</svg>
	<figcaption>
		<span class:shown={done}>{demo.caption}</span>
		<button type="button" class="textlink" onclick={play}>Replay</button>
	</figcaption>
</figure>

<style>
	.demo {
		margin: 0;
	}
	svg {
		display: block;
		width: 100%;
		height: auto;
	}
	figcaption {
		display: flex;
		justify-content: space-between;
		align-items: baseline;
		gap: var(--s2);
		padding-top: var(--s1);
		font-size: var(--step--1);
		color: var(--text-soft);
	}
	figcaption span {
		opacity: 0;
		transition: opacity 400ms;
	}
	figcaption span.shown {
		opacity: 1;
	}
</style>
