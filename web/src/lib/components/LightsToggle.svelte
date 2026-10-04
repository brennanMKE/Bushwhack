<script lang="ts">
	import { lights, restoreLights, setLights } from '#lib/lights.svelte.ts';

	$effect(restoreLights);
</script>

<button class="lights" type="button" aria-pressed={lights.on} onclick={() => setLights(!lights.on)}>
	<svg viewBox="0 0 16 16" aria-hidden="true"><path d="M8 1.5c1.8 2.4 3 4 3 6a3 3 0 0 1-6 0c0-1.2.6-2.1 1.3-3 .2 1 .7 1.6 1.3 1.8C7.3 4.8 7.5 3 8 1.5Z" /><rect x="6.5" y="11.5" width="3" height="3" rx="0.5" /></svg>
	{lights.on ? 'Lights on' : 'Lights off'}
</button>

<style>
	.lights {
		display: inline-flex;
		align-items: center;
		gap: 0.4rem;
		font: inherit;
		font-weight: 700;
		font-size: var(--step--1);
		padding: 0.35rem 0.75rem;
		border: 1.5px solid var(--line);
		border-radius: var(--radius);
		background: transparent;
		color: var(--text);
		cursor: pointer;
	}
	.lights[aria-pressed='true'] {
		border-color: var(--brass);
		background: color-mix(in srgb, var(--brass) 14%, transparent);
	}
	svg {
		width: 1rem;
		height: 1rem;
		fill: currentColor;
	}
	.lights[aria-pressed='true'] svg {
		fill: #e07b1f;
	}
</style>
