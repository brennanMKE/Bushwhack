<script lang="ts">
	import type { Pattern } from '#lib/api.ts';
	import Swatch from './Swatch.svelte';
	import { lights, restoreLights } from '#lib/lights.svelte.ts';

	$effect(restoreLights);

	let { pattern, href = `/patterns/${pattern.slug}` }: { pattern: Pattern; href?: string } = $props();
</script>

<a class="tile" {href}>
	<Swatch drawing={pattern.drawing} widthIn={pattern.widthIn} heightIn={pattern.heightIn} label="{pattern.name} pattern" lit={lights.on} />
	<span>{pattern.name}</span>
</a>

<style>
	.tile {
		display: grid;
		gap: 0.45rem;
		text-decoration: none;
		font-weight: 600;
	}
	.tile:hover span {
		text-decoration: underline;
		text-decoration-color: var(--brass);
		text-decoration-thickness: 2px;
		text-underline-offset: 3px;
	}
</style>
