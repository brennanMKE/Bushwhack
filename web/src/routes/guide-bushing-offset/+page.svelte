<script lang="ts">
	import Seo from '#lib/components/Seo.svelte';
	import MeasureInput from '#lib/components/MeasureInput.svelte';
	import OffsetTable from '#lib/components/OffsetTable.svelte';
	import { holeOffset, pieceOffset } from '#lib/offsets.ts';
	import { parseLength, fraction } from '#lib/units.ts';

	let bushing = $state('5/16');
	let bit = $state('1/8');

	const b = $derived(parseLength(bushing));
	const d = $derived(parseLength(bit));
	const ok = $derived(!!b && !!d && b.inches > 0 && d.inches > 0 && d.inches < b.inches);
	const hole = $derived(ok ? holeOffset(b!.inches, d!.inches) : 0);
	const piece = $derived(ok ? pieceOffset(b!.inches, d!.inches) : 0);

	// Fraction when it's a clean 64th, then decimal inches and millimetres.
	const show = (inches: number) => {
		const f = fraction(inches);
		const parts = f.includes('.') ? [] : [f];
		parts.push(`${inches.toFixed(4)} in`, `${(inches * 25.4).toFixed(2)} mm`);
		return parts;
	};
	const makeHref = $derived(
		'/make?' + new URLSearchParams({ bushing, bit }).toString().replace(/%2F/gi, '/')
	);
</script>

<Seo
	title="Guide bushing offset calculator · Bushwhack"
	description="Calculate the router template offset for any guide bushing and bit: (bushing OD − bit diameter) ÷ 2 for a hole, (bushing OD + bit diameter) ÷ 2 for a piece or inlay."
	path="/guide-bushing-offset"
/>

<article class="wrap page">
	<h1>Guide bushing offset calculator</h1>
	<p class="prose lead">Enter your bushing's outside diameter and your bit. Inches, fractions or millimetres all work: 5/16, 0.3125, 8mm.</p>

	<div class="calc">
		<div class="inputs">
			<MeasureInput id="bushing" label="Bushing outside diameter" bind:value={bushing} suggestions={['5/16', '3/8', '7/16', '1/2', '5/8', '3/4', '1', '10mm', '17mm', '30mm']} />
			<MeasureInput id="bit" label="Bit diameter" bind:value={bit} suggestions={['1/8', '3/16', '1/4', '5/16', '3/8', '1/2', '3mm', '6mm', '8mm']} />
		</div>
		<div class="results" aria-live="polite">
			{#if ok}
				<div class="result">
					<h2>Hole offset</h2>
					<p class="big">{show(hole)[0]}</p>
					<p class="alt">{show(hole).slice(1).join(' · ')}</p>
					<p class="small">The opening in your board matches the drawing. Pockets, recesses, cut-outs.</p>
				</div>
				<div class="result">
					<h2>Piece offset</h2>
					<p class="big">{show(piece)[0]}</p>
					<p class="alt">{show(piece).slice(1).join(' · ')}</p>
					<p class="small">The part that falls out matches the drawing. Inlays, letters.</p>
				</div>
			{:else}
				<p class="note">The bit must be smaller than the bushing. Check both sizes.</p>
			{/if}
		</div>
	</div>
	{#if ok}
		<p><a class="btn" href={makeHref}>Make a template with this setup</a></p>
	{/if}

	<section aria-labelledby="how">
		<h2 id="how">How the offset works</h2>
		<div class="prose">
			<p>The bushing's outside wall rides the template and the bit spins in its centre. The bit's near edge is the bushing radius minus the bit radius from the template; its far edge is the bushing radius plus the bit radius. So:</p>
			<ul>
				<li><strong>Hole offset</strong> = (bushing OD − bit diameter) ÷ 2. Grow each template opening by this much and the hole you rout matches your drawing.</li>
				<li><strong>Piece offset</strong> = (bushing OD + bit diameter) ÷ 2. Grow each opening by this much and the piece that falls out matches your drawing.</li>
			</ul>
			<p>The two always differ by exactly one bit diameter. That's why inlay kits pair a bushing with a removable collar: the collar adds the bit diameter to the bushing.</p>
			<p>Growing an opening isn't the same as scaling it. Scaling a 4 in star up to 4 3/16 in widens its body by 3/16 in but its narrow points by only a few thousandths, so the points still come out too thin. An offset moves every edge out by the same distance, which is what the bushing does.</p>
			<p>Bushwhack does the offset for a whole drawing: upload an SVG, enter the same two sizes, and download a true-size template.</p>
		</div>
		<OffsetTable caption="Offsets for common bushing and bit pairs" />
	</section>

	<section aria-labelledby="check">
		<h2 id="check">Before you rout</h2>
		<ul class="prose">
			<li>The bit has to pass through the bushing's <em>inside</em> diameter with room to spare. Check it by hand, unplugged.</li>
			<li>Measure the bushing with calipers. Nominal sizes are sometimes a few thousandths off, and the offset changes by half of any error.</li>
			<li>Centre the bushing on the bit, or the offset will differ on each side. A centring cone or pin does this.</li>
		</ul>
		<p>More in the <a href="/guide">guide</a>, and on <a href="/inlays">inlay templates</a>.</p>
	</section>
</article>

<style>
	.page {
		padding-block: var(--s6) var(--s8);
	}
	.lead {
		font-size: var(--step-1);
		color: var(--text-soft);
	}
	.calc {
		display: grid;
		grid-template-columns: minmax(0, 20rem) minmax(0, 1fr);
		gap: var(--s4);
		align-items: start;
		margin-block: var(--s3);
	}
	.inputs {
		display: grid;
		gap: var(--s2);
	}
	.results {
		display: grid;
		grid-template-columns: repeat(2, minmax(0, 1fr));
		gap: var(--s3);
	}
	.result {
		border-left: 4px solid var(--brass);
		padding-left: var(--s2);
	}
	.result h2 {
		font-size: var(--step-2);
		margin-bottom: var(--s1);
	}
	.big {
		font-size: var(--step-3);
		font-weight: 800;
		line-height: 1.1;
		margin-bottom: 0.2rem;
	}
	.alt,
	.small {
		font-size: var(--step--1);
		color: var(--text-soft);
	}
	section {
		margin-top: var(--s6);
	}
	section h2 {
		font-size: var(--step-3);
	}
	li {
		margin-bottom: var(--s1);
	}
	@media (max-width: 800px) {
		.calc,
		.results {
			grid-template-columns: 1fr;
		}
	}
</style>
