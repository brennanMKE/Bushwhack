<script lang="ts">
	// Bushing and bit fields with optional pick lists of known routers,
	// bushings and bits. Picking fills the fields; typing still works, and the
	// lists fall back to "type a size" when the fields no longer match.
	import MeasureInput from './MeasureInput.svelte';
	import { bits, innerFor, pickedBit, pickedBushing, routers, setsFor } from '#lib/gear.ts';
	import { fraction, parseLength } from '#lib/units.ts';

	let {
		bushing = $bindable(),
		inner = $bindable(),
		bit = $bindable(),
		router = $bindable(),
		bushingPick = $bindable(),
		bitPick = $bindable()
	}: { bushing: string; inner: string; bit: string; router: string; bushingPick: string; bitPick: string } = $props();

	const theRouter = $derived(routers.find((r) => r.id === router) ?? null);
	const sets = $derived(setsFor(router));
	const bushingSel = $derived(pickedBushing(bushingPick, bushing)?.bushing.id ?? '');
	const bitSel = $derived(pickedBit(bitPick, bit)?.id ?? '');
	const innerNow = $derived(innerFor(bushing, inner, bushingPick));

	const size = (s: string | null) => {
		const p = s ? parseLength(s) : null;
		if (!p) return '?';
		return p.unit === 'in' ? fraction(p.inches).replace(' in', '') : s!;
	};
	const bitGroups = $derived(
		[...new Set(bits.map((b) => b.dia))].map((dia) => ({ dia, items: bits.filter((b) => b.dia === dia) }))
	);

	function pickBushing(id: string) {
		bushingPick = id;
		for (const s of sets) {
			const b = s.bushings.find((x) => x.id === id);
			if (b) {
				bushing = b.od;
				inner = '';
			}
		}
	}
	function pickBit(id: string) {
		bitPick = id;
		const b = bits.find((x) => x.id === id);
		if (b) bit = b.dia;
	}
</script>

<div class="gear">
	<div class="pick">
		<label for="router">Router <span class="opt">optional</span></label>
		<select id="router" bind:value={router}>
			<option value="">Any router</option>
			{#each routers as r (r.id)}<option value={r.id}>{r.name}</option>{/each}
		</select>
		{#if theRouter}<p class="note-sm">{theRouter.note}</p>{/if}
	</div>

	{#if sets.length}
		<div class="pick">
			<label for="bushing-pick">Guide bushing</label>
			<select id="bushing-pick" value={bushingSel} onchange={(e) => pickBushing(e.currentTarget.value)}>
				<option value="">Type a size below</option>
				{#each sets as s (s.id)}
					<optgroup label={s.name}>
						{#each s.bushings as b (b.id)}
							<option value={b.id}>{b.part === b.od || b.part === size(b.od) ? '' : b.part + ': '}{size(b.od)} OD, {size(b.id_)} ID</option>
						{/each}
					</optgroup>
				{/each}
			</select>
		</div>
	{/if}

	<div class="pair">
		<MeasureInput
			id="bushing"
			label="Bushing OD"
			bind:value={() => bushing, (v) => ((bushing = v), (inner = ''))}
			suggestions={['5/16', '3/8', '7/16', '1/2', '5/8', '3/4', '1', '10mm', '17mm', '30mm']}
		/>
		<MeasureInput
			id="bushing-inner"
			label="Bushing ID"
			bind:value={() => innerNow.value, (v) => (inner = v)}
			hint={innerNow.from}
		/>
	</div>

	<div class="pick">
		<label for="bit-pick">Bit</label>
		<select id="bit-pick" value={bitSel} onchange={(e) => pickBit(e.currentTarget.value)}>
			<option value="">Type a size below</option>
			{#each bitGroups as g (g.dia)}
				<optgroup label="{size(g.dia)} in">
					{#each g.items as b (b.id)}
						<option value={b.id}>{b.name}, {b.kind.toLowerCase()}{b.cut ? `, ${size(b.cut)} in cut` : ''}</option>
					{/each}
				</optgroup>
			{/each}
		</select>
	</div>
	<MeasureInput
		id="bit"
		label="Bit diameter"
		bind:value={bit}
		suggestions={['1/8', '3/16', '1/4', '5/16', '3/8', '1/2', '3mm', '6mm', '8mm']}
	/>
	<p class="note-sm"><a href="/gear">About these routers, bushings and bits</a></p>
</div>

<style>
	.gear {
		display: grid;
		gap: var(--s2);
	}
	.pick {
		display: grid;
		gap: 0.3rem;
	}
	label {
		font-weight: 700;
	}
	.opt {
		font-weight: 400;
		color: var(--text-soft);
		font-size: var(--step--1);
	}
	select {
		font: inherit;
		width: 100%;
		padding: 0.45rem 0.55rem;
		border: 1.5px solid var(--line);
		border-radius: var(--radius);
		background: var(--surface);
		color: var(--text);
	}
	/* Subgrid rows keep both inputs level even when one label or note wraps. */
	.pair {
		display: grid;
		grid-template-columns: repeat(2, minmax(0, 1fr));
		grid-template-rows: auto auto auto;
		column-gap: var(--s2);
		row-gap: 0.3rem;
	}
	.pair > :global(.measure) {
		grid-row: span 3;
		grid-template-rows: subgrid;
	}
	.pair > :global(.measure > label) {
		align-self: end;
	}
	.note-sm {
		margin: 0;
		font-size: var(--step--1);
		color: var(--text-soft);
	}
</style>
