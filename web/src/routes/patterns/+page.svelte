<script lang="ts">
	import { onMount } from 'svelte';
	import PatternTile from '#lib/components/PatternTile.svelte';
	import { loadPatterns, type Occasion } from '#lib/api.ts';

	let occasions = $state<Occasion[]>([]);
	let failed = $state(false);
	onMount(async () => {
		try {
			occasions = (await loadPatterns()).filter((o) => o.patterns.length);
		} catch {
			failed = true;
		}
	});
</script>

<svelte:head><title>Patterns · Bushwhack</title></svelte:head>

<div class="wrap page">
	<h1>Patterns</h1>
	<p class="prose lead">Original starter shapes for holidays and special days, each checked to cut cleanly with a 5/16 in bushing and 1/8 in bit at its recommended size. Free to use, sell what you make.</p>

	{#if failed}
		<div class="note" role="alert"><span>Patterns didn't load. Refresh the page to try again.</span></div>
	{/if}

	{#each occasions as o (o.slug)}
		<section aria-labelledby="o-{o.slug}">
			<h2 id="o-{o.slug}">{o.name}{#if o.upcoming}<span class="soon"> coming up</span>{/if}</h2>
			<p class="prose desc">{o.description}</p>
			<div class="tiles">
				{#each o.patterns as p (p.slug)}
					<PatternTile pattern={p} />
				{/each}
			</div>
		</section>
	{/each}
</div>

<style>
	.page {
		padding-block: var(--s6) var(--s8);
	}
	.lead {
		font-size: var(--step-1);
		color: var(--text-soft);
		margin-bottom: var(--s6);
	}
	section {
		margin-bottom: var(--s6);
	}
	.soon {
		font-family: var(--font-text);
		font-size: var(--step-0);
		font-weight: 700;
		color: var(--text-soft);
		vertical-align: middle;
		margin-left: 0.5rem;
	}
	.desc {
		color: var(--text-soft);
	}
	.tiles {
		display: grid;
		grid-template-columns: repeat(auto-fill, minmax(150px, 1fr));
		gap: var(--s3);
	}
</style>
