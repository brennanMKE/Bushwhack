<script lang="ts">
	import { combos, holeOffset, pieceOffset } from '#lib/offsets.ts';
	import { fraction } from '#lib/units.ts';

	let { caption }: { caption: string } = $props();

	// Metric pairs read better in millimetres.
	const show = (inches: number, metric: boolean) =>
		metric ? `${+(inches * 25.4).toFixed(2)} mm` : fraction(inches);
</script>

<div class="scroll">
	<table>
		<caption>{caption}</caption>
		<thead>
			<tr><th scope="col">Bushing OD</th><th scope="col">Bit</th><th scope="col">Hole offset</th><th scope="col">Piece offset</th></tr>
		</thead>
		<tbody>
			{#each combos as c (c.bushing + c.bit)}
				{@const metric = c.bushing.endsWith('mm')}
				<tr>
					<td>{c.bushing}</td>
					<td>{c.bit}</td>
					<td>{show(holeOffset(c.bushingIn, c.bitIn), metric)}</td>
					<td>{show(pieceOffset(c.bushingIn, c.bitIn), metric)}</td>
				</tr>
			{/each}
		</tbody>
	</table>
</div>

<style>
	.scroll {
		overflow-x: auto;
		margin-bottom: var(--s3);
	}
	table {
		border-collapse: collapse;
		min-width: 24rem;
	}
	caption {
		text-align: left;
		font-size: var(--step--1);
		color: var(--text-soft);
		padding-bottom: var(--s1);
	}
	th,
	td {
		text-align: left;
		padding: 0.4rem 1.2rem 0.4rem 0;
		border-bottom: 1px solid var(--rule);
	}
	th {
		font-weight: 700;
	}
</style>
