<script lang="ts">
	import { onMount } from 'svelte';
	import HeroDemo from '#lib/components/HeroDemo.svelte';
	import PatternTile from '#lib/components/PatternTile.svelte';
	import { loadPatterns, type Occasion } from '#lib/api.ts';

	let upcoming = $state<Occasion | null>(null);
	onMount(async () => {
		try {
			const all = await loadPatterns();
			upcoming = all.find((o) => o.upcoming && o.patterns.length) ?? all.find((o) => o.patterns.length) ?? null;
		} catch {
			upcoming = null;
		}
	});
</script>

<svelte:head><title>Bushwhack: router templates that cut where you drew</title></svelte:head>

<section class="wrap hero">
	<div class="copy">
		<h1>Templates that cut where you drew.</h1>
		<p class="sub">Upload an SVG, tell us your guide bushing and bit, and get a template sized so your router lands right on the line.</p>
		<div class="actions">
			<a class="btn" href="/make">Make a template</a>
			<a class="textlink" href="/patterns">Browse patterns</a>
		</div>
	</div>
	<div class="stage"><HeroDemo /></div>
</section>

<section class="wrap how" aria-labelledby="how">
	<h2 id="how">How it works</h2>
	<ol>
		<li><strong>Upload</strong> an SVG of the shapes you want to cut, or start from a pattern.</li>
		<li><strong>Set your bushing and bit.</strong> Bushwhack grows every opening by exactly the right amount.</li>
		<li><strong>Download</strong> a true-size template for your laser, CNC or scroll saw, and rout.</li>
	</ol>
</section>

{#if upcoming}
	<section class="wrap season" aria-labelledby="season">
		<h2 id="season">Coming up: {upcoming.name}</h2>
		<p class="prose">{upcoming.description}</p>
		<div class="tiles">
			{#each upcoming.patterns.slice(0, 4) as p (p.slug)}
				<PatternTile pattern={p} />
			{/each}
		</div>
		<p><a href="/patterns">See all patterns</a></p>
	</section>
{/if}

<style>
	.hero {
		display: grid;
		grid-template-columns: 5fr 7fr;
		gap: var(--s6);
		align-items: center;
		padding-block: var(--s6) var(--s8);
	}
	.sub {
		font-size: var(--step-1);
		max-width: 34ch;
		color: var(--text-soft);
	}
	.actions {
		display: flex;
		align-items: center;
		gap: var(--s3);
		flex-wrap: wrap;
		margin-top: var(--s3);
	}
	.how {
		padding-block: var(--s4);
		border-top: 1px solid var(--rule);
	}
	.how ol {
		display: grid;
		grid-template-columns: repeat(3, 1fr);
		gap: var(--s4);
		padding: 0;
		margin: 0;
		list-style: none;
		counter-reset: step;
	}
	.how li {
		counter-increment: step;
		max-width: 32ch;
	}
	.how li::before {
		content: counter(step);
		display: block;
		font-family: var(--font-display);
		font-weight: 800;
		font-size: var(--step-4);
		line-height: 1;
		color: var(--brass);
		margin-bottom: var(--s1);
	}
	.season {
		padding-block: var(--s4) var(--s8);
		border-top: 1px solid var(--rule);
	}
	.tiles {
		display: grid;
		grid-template-columns: repeat(4, 1fr);
		gap: var(--s3);
		margin-bottom: var(--s3);
	}
	@media (max-width: 860px) {
		.hero {
			grid-template-columns: 1fr;
			gap: var(--s3);
			padding-top: var(--s3);
		}
		.stage {
			order: -1;
		}
		.how ol {
			grid-template-columns: 1fr;
			gap: var(--s3);
		}
		.tiles {
			grid-template-columns: repeat(2, 1fr);
		}
	}
</style>
