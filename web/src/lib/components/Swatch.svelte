<script lang="ts">
	// A square hardboard swatch with the pattern's openings cut through it, so
	// the mat shows behind. Lit, it becomes the front of a light box: a candle
	// sits at the bottom centre behind it and glows through the openings.
	import { light, STILL, reducedMotion, type FlameState } from '#lib/flicker.ts';

	let {
		drawing,
		widthIn,
		heightIn,
		label = '',
		lit = false
	}: { drawing: string; widthIn: number; heightIn: number; label?: string; lit?: boolean } = $props();

	const uid = $props.id();
	const side = $derived(Math.max(widthIn, heightIn));
	const x = $derived((widthIn - side) / 2);
	const y = $derived((heightIn - side) / 2);
	const sheet = $derived(`M${x} ${y}h${side}v${side}h${-side}Z`);

	// The candle, in drawing units: bottom centre, below the openings.
	const cx = $derived(x + side / 2);
	const cy = $derived(y + side * 0.97);

	let svg = $state<SVGSVGElement>();
	let flameEl = $state<SVGRadialGradientElement>();
	let hot = $state<SVGRectElement>();
	let glow = $state<SVGGElement>();

	function draw(s: FlameState) {
		if (!flameEl || !hot || !glow) return;
		// A brighter flame burns taller and yellower; a dim one shrinks to orange.
		flameEl.setAttribute('cx', String(cx + s.sway * side * 0.04));
		flameEl.setAttribute('cy', String(cy - (s.k - 0.8) * side * 0.15));
		flameEl.setAttribute('r', String(side * (0.85 + 0.35 * s.k)));
		hot.setAttribute('opacity', (s.k * s.k).toFixed(3));
		glow.setAttribute('opacity', (0.15 + 0.35 * s.k).toFixed(3));
	}

	// Flicker only while the swatch is on screen.
	$effect(() => {
		if (!lit || !svg || reducedMotion()) return;
		let stop: (() => void) | null = null;
		const io = new IntersectionObserver(([e]) => {
			if (e.isIntersecting && !stop) stop = light(draw);
			else if (!e.isIntersecting && stop) {
				stop();
				stop = null;
			}
		});
		io.observe(svg);
		return () => {
			io.disconnect();
			stop?.();
		};
	});
</script>

<svg bind:this={svg} class="swatch" viewBox="{x} {y} {side} {side}" role="img" aria-label={label}>
	{#if lit}
		<defs>
			<!-- Ember: the dim, orange light that's always there. -->
			<radialGradient id="{uid}-ember" gradientUnits="userSpaceOnUse" {cx} {cy} r={side * 1.25}>
				<stop offset="0" stop-color="#ff9a3c" />
				<stop offset="0.5" stop-color="#d9551c" />
				<stop offset="1" stop-color="#6e1f0a" />
			</radialGradient>
			<!-- Flame: the bright core, moved and dimmed every frame. -->
			<radialGradient bind:this={flameEl} id="{uid}-flame" gradientUnits="userSpaceOnUse" {cx} cy={cy - side * 0.015} r={side * (0.85 + 0.35 * STILL.k)}>
				<stop offset="0" stop-color="#fff8dc" />
				<stop offset="0.3" stop-color="#ffd27a" />
				<stop offset="0.75" stop-color="#ff9a3c" stop-opacity="0.6" />
				<stop offset="1" stop-color="#ff9a3c" stop-opacity="0" />
			</radialGradient>
			<filter id="{uid}-blur" x="-20%" y="-20%" width="140%" height="140%">
				<feGaussianBlur stdDeviation={side * 0.03} />
			</filter>
		</defs>
		<rect {x} {y} width={side} height={side} fill="url(#{uid}-ember)" />
		<rect bind:this={hot} {x} {y} width={side} height={side} fill="url(#{uid}-flame)" opacity={STILL.k * STILL.k} />
		<path d="{sheet}{drawing}" fill="#4a301b" fill-rule="evenodd" />
		<!-- Light spilling past the edges of each opening. -->
		<g bind:this={glow} class="glow" opacity={0.15 + 0.35 * STILL.k}>
			<path d={drawing} fill="#ffb347" filter="url(#{uid}-blur)" />
		</g>
	{:else}
		<rect {x} {y} width={side} height={side} fill="var(--mat)" />
		<path d="{sheet}{drawing}" fill="var(--hardboard)" fill-rule="evenodd" />
	{/if}
</svg>

<style>
	.swatch {
		display: block;
		width: 100%;
		height: auto;
		aspect-ratio: 1;
	}
	.glow {
		mix-blend-mode: screen;
		pointer-events: none;
	}
</style>
