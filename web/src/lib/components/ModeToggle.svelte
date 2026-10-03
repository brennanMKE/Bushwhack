<script lang="ts">
	let { mode = $bindable() }: { mode: 'hole' | 'piece' } = $props();

	const options = [
		{ value: 'hole', label: 'Cut a hole', help: 'The opening matches your drawing.' },
		{ value: 'piece', label: 'Cut out a piece', help: 'The part that falls out matches your drawing.' }
	] as const;
	const help = $derived(options.find((o) => o.value === mode)?.help ?? '');
</script>

<fieldset class="mode">
	<legend>What should match your drawing?</legend>
	<div class="pill" role="radiogroup" aria-label="What should match your drawing?">
		{#each options as o (o.value)}
			<label class:on={mode === o.value}>
				<input type="radio" name="mode" value={o.value} bind:group={mode} />
				{o.label}
			</label>
		{/each}
	</div>
	<div class="explain">
		<!-- Which side of the red cut is kept: the hardboard-coloured part. -->
		<svg viewBox="0 0 120 44" aria-hidden="true">
			{#if mode === 'hole'}
				<rect x="2" y="4" width="116" height="36" fill="var(--hardboard)" />
				<rect x="34" y="12" width="52" height="20" fill="var(--bg)" />
				<rect x="34" y="12" width="52" height="20" fill="none" stroke="var(--cut)" stroke-width="2.5" />
				<rect x="34" y="12" width="52" height="20" fill="none" stroke="var(--chalk)" stroke-width="1.5" stroke-dasharray="4 3" transform="translate(0 0)" />
			{:else}
				<rect x="2" y="4" width="116" height="36" fill="none" stroke="var(--line)" stroke-dasharray="3 3" />
				<rect x="34" y="12" width="52" height="20" fill="var(--hardboard)" />
				<rect x="34" y="12" width="52" height="20" fill="none" stroke="var(--cut)" stroke-width="2.5" />
				<rect x="34" y="12" width="52" height="20" fill="none" stroke="var(--chalk)" stroke-width="1.5" stroke-dasharray="4 3" />
			{/if}
		</svg>
		<p>{help} {mode === 'hole' ? 'You keep the board around it.' : 'You keep the piece; the board is scrap.'}</p>
	</div>
</fieldset>

<style>
	fieldset {
		border: 0;
		padding: 0;
		margin: 0;
		min-width: 0;
	}
	legend {
		font-weight: 700;
		margin-bottom: 0.4rem;
		padding: 0;
	}
	.pill {
		display: grid;
		grid-template-columns: 1fr 1fr;
		border: 1.5px solid var(--line);
		border-radius: 999px;
		padding: 3px;
		background: var(--surface);
	}
	label {
		position: relative;
		text-align: center;
		padding: 0.45rem 0.5rem;
		border-radius: 999px;
		cursor: pointer;
		font-weight: 600;
		line-height: 1.2;
	}
	label.on {
		background: var(--brass);
		color: var(--graphite);
	}
	label:has(input:focus-visible) {
		outline: 2px solid var(--focus);
		outline-offset: 2px;
	}
	input {
		position: absolute;
		opacity: 0;
		inset: 0;
		margin: 0;
		cursor: pointer;
	}
	.explain {
		display: grid;
		grid-template-columns: 5.5rem 1fr;
		gap: 0.75rem;
		align-items: center;
		margin-top: 0.6rem;
	}
	.explain svg {
		width: 100%;
		height: auto;
	}
	.explain p {
		margin: 0;
		font-size: var(--step--1);
		color: var(--text-soft);
	}
</style>
