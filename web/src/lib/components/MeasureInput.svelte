<script lang="ts">
	import { parseLength, describe } from '#lib/units.ts';

	let {
		id,
		label,
		value = $bindable(),
		suggestions = [],
		hint = ''
	}: { id: string; label: string; value: string; suggestions?: string[]; hint?: string } = $props();

	const parsed = $derived(parseLength(value));
	const valid = $derived(parsed !== null && parsed.inches > 0);
</script>

<div class="measure">
	<label for={id}>{label}</label>
	<input
		{id}
		bind:value
		list={suggestions.length ? `${id}-list` : undefined}
		inputmode="decimal"
		autocomplete="off"
		spellcheck="false"
		aria-invalid={!valid}
		aria-describedby="{id}-echo"
	/>
	{#if suggestions.length}
		<datalist id="{id}-list">
			{#each suggestions as s (s)}<option value={s}></option>{/each}
		</datalist>
	{/if}
	<p class="echo" class:bad={!valid} id="{id}-echo">
		{#if valid && parsed}{describe(parsed)}{#if hint}<span class="hint"> · {hint}</span>{/if}{:else}Enter a size like 5/16 or 0.3125.{/if}
	</p>
</div>

<style>
	.measure {
		display: grid;
		gap: 0.3rem;
	}
	label {
		font-weight: 700;
	}
	input {
		font: inherit;
		font-size: var(--step-1);
		font-weight: 600;
		width: 100%;
		padding: 0.45rem 0.65rem;
		border: 1.5px solid var(--line);
		border-radius: var(--radius);
		background: var(--surface);
		color: var(--text);
	}
	input[aria-invalid='true'] {
		border-color: var(--brass);
		border-width: 2px;
	}
	.echo {
		margin: 0;
		font-size: var(--step--1);
		color: var(--text-soft);
		min-height: 1.3em;
	}
	.echo.bad {
		color: var(--text);
		border-left: 3px solid var(--brass);
		padding-left: 0.45rem;
	}
	.hint {
		color: var(--text-soft);
	}
</style>
