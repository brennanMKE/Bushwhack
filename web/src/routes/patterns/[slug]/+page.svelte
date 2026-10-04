<script lang="ts">
	import { page } from '$app/state';
	import Swatch from '#lib/components/Swatch.svelte';
	import LightsToggle from '#lib/components/LightsToggle.svelte';
	import { lights } from '#lib/lights.svelte.ts';
	import { loadPatterns, type Occasion, type Pattern } from '#lib/api.ts';
	import { fraction } from '#lib/units.ts';

	let pattern = $state<Pattern | null>(null);
	let occasion = $state<Occasion | null>(null);
	let missing = $state(false);

	$effect(() => {
		const slug = page.params.slug;
		loadPatterns()
			.then((all) => {
				for (const o of all) {
					const p = o.patterns.find((x) => x.slug === slug);
					if (p) {
						pattern = p;
						occasion = o;
						missing = false;
						return;
					}
				}
				missing = true;
			})
			.catch(() => (missing = true));
	});
</script>

<svelte:head><title>{pattern ? `${pattern.name} pattern` : 'Pattern'} · Bushwhack</title></svelte:head>

<div class="wrap page">
	{#if missing}
		<h1>Pattern not found</h1>
		<p><a href="/patterns">See all patterns</a></p>
	{:else if pattern}
		<p class="crumb"><a href="/patterns">Patterns</a> / {occasion?.name}</p>
		<div class="detail">
			<div class="big">
				<Swatch drawing={pattern.drawing} widthIn={pattern.widthIn} heightIn={pattern.heightIn} label="{pattern.name} pattern at its recommended size" lit={lights.on} />
				<p class="toggle"><LightsToggle /></p>
			</div>
			<div class="info">
				<h1>{pattern.name}</h1>
				<dl>
					<div><dt>Recommended size</dt><dd>{fraction(pattern.recommendedSizeIn)} longest side</dd></div>
					<div>
						<dt>Smallest safe size</dt>
						<dd>{pattern.minSafeSizeIn ? `${fraction(pattern.minSafeSizeIn)}` : 'still measuring…'}</dd>
					</div>
				</dl>
				<p class="small">Sizes are for a 5/16 in bushing and a 1/8 in bit. Below the smallest safe size, openings get too close together or lose detail.</p>
				<p>{pattern.note}</p>
				<div class="actions">
					<a class="btn" href="/make?pattern={pattern.slug}">Open in workbench</a>
					<a class="textlink" href="/api/patterns/{pattern.slug}.svg" download>Download SVG</a>
				</div>
				<p class="license">Original to Bushwhack and released as CC0. Free to use, sell what you make.</p>
			</div>
		</div>
	{/if}
</div>

<style>
	.page {
		padding-block: var(--s4) var(--s8);
	}
	.toggle {
		margin-top: var(--s1);
	}
	.crumb {
		font-size: var(--step--1);
		color: var(--text-soft);
	}
	.detail {
		display: grid;
		grid-template-columns: 7fr 5fr;
		gap: var(--s6);
		align-items: start;
	}
	dl {
		display: grid;
		gap: var(--s1);
		margin: 0 0 var(--s2);
	}
	dt {
		font-size: var(--step--1);
		color: var(--text-soft);
	}
	dd {
		margin: 0;
		font-weight: 700;
		font-size: var(--step-1);
	}
	.small,
	.license {
		font-size: var(--step--1);
		color: var(--text-soft);
	}
	.actions {
		display: flex;
		gap: var(--s3);
		align-items: center;
		flex-wrap: wrap;
		margin: var(--s3) 0;
	}
	@media (max-width: 800px) {
		.detail {
			grid-template-columns: 1fr;
			gap: var(--s3);
		}
	}
</style>
