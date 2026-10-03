<script lang="ts">
	// A square hardboard swatch with the pattern's openings cut through it, so
	// the mat shows behind.
	let {
		drawing,
		widthIn,
		heightIn,
		label = ''
	}: { drawing: string; widthIn: number; heightIn: number; label?: string } = $props();

	const side = $derived(Math.max(widthIn, heightIn));
	const x = $derived((widthIn - side) / 2);
	const y = $derived((heightIn - side) / 2);
	const sheet = $derived(`M${x} ${y}h${side}v${side}h${-side}Z`);
</script>

<svg class="swatch" viewBox="{x} {y} {side} {side}" role="img" aria-label={label}>
	<rect {x} {y} width={side} height={side} fill="var(--mat)" />
	<path d="{sheet}{drawing}" fill="var(--hardboard)" fill-rule="evenodd" />
</svg>

<style>
	.swatch {
		display: block;
		width: 100%;
		height: auto;
		aspect-ratio: 1;
	}
</style>
