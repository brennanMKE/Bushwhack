<script lang="ts">
	import Seo from '#lib/components/Seo.svelte';
	import { bits, bushingSets, routers, systems, standardInner, type System } from '#lib/gear.ts';
	import { fraction, parseLength } from '#lib/units.ts';

	// "5/16" → "5/16 in", "17mm" → "17 mm", null → "—".
	const size = (s: string | null) => {
		const p = s ? parseLength(s) : null;
		if (!p) return '—';
		if (p.unit === 'mm') return `${+p.value.toFixed(1)} mm`;
		return fraction(p.inches);
	};
	const shortSystem: Record<System, string> = {
		pc: 'Porter-Cable style',
		bosch: 'Bosch quick-change',
		festool: 'Festool',
		triton: 'Triton'
	};
	const bitSizes = [1 / 8, 3 / 16, 1 / 4, 5 / 16, 3 / 8, 1 / 2];
	const standardODs = ['5/16', '3/8', '7/16', '1/2', '5/8', '3/4', '51/64', '1', '17mm'];
</script>

<Seo
	title="Router guide bushings and bits: sizes for template routing · Bushwhack"
	description="Guide bushing outside and inside diameters, which bushings fit common routers, and bits that pass through them, for template routing with a guide bushing."
	path="/gear"
/>

<article class="wrap page">
	<h1>Routers, bushings and bits</h1>
	<p class="prose lead">The sizes Bushwhack needs: each bushing's outside diameter, which sets the offset, and its inside diameter, which the bit has to pass through. You can pick any of these on the <a href="/make">workbench</a>.</p>
	<p class="prose">Sizes come from makers' and distributors' published specs. A dash means we couldn't confirm the number. Bushings vary by a few thousandths, so measure yours with calipers before you rely on a template.</p>

	<section aria-labelledby="standard">
		<h2 id="standard">Standard bushing sizes</h2>
		<p class="prose">Most bushings follow the Porter-Cable sizes. When you type an outside diameter on the workbench, Bushwhack assumes the matching inside diameter here unless you enter your own, and it won't make a template if the bit can't pass through.</p>
		<div class="scroll">
			<table>
				<thead><tr><th scope="col">Outside</th><th scope="col">Inside</th><th scope="col">Bits that fit</th></tr></thead>
				<tbody>
					{#each standardODs as od (od)}
						{@const id = standardInner(od)}
						{@const idIn = parseLength(id ?? '')?.inches ?? 0}
						<tr>
							<td>{size(od)}</td>
							<td>{size(id)}</td>
							<td>{bitSizes.filter((b) => b < idIn).map((b) => fraction(b).replace(' in', '')).join(', ')} in</td>
						</tr>
					{/each}
				</tbody>
			</table>
		</div>
		<p class="prose small">Common pairings: a 5/16 in bushing with a 1/8 in bit for fine detail, and 7/16 in or larger with a 1/4 in bit for bigger shapes.</p>
	</section>

	<section aria-labelledby="routers">
		<h2 id="routers">Which bushings fit your router</h2>
		<p class="prose">Bushings attach in a few ways. The most common is the Porter-Cable style: {systems.pc.split(': ')[1]}.</p>
		<div class="scroll">
			<table>
				<thead><tr><th scope="col">Router</th><th scope="col">Type</th><th scope="col">Collets</th><th scope="col">Bushings</th><th scope="col">Notes</th></tr></thead>
				<tbody>
					{#each routers as r (r.id)}
						<tr>
							<th scope="row">{r.name}</th>
							<td>{r.kind}</td>
							<td>{r.collets}</td>
							<td>{r.takes.length ? r.takes.map((t) => shortSystem[t]).join(', ') : 'None as shipped'}</td>
							<td class="note-cell">{r.note}</td>
						</tr>
					{/each}
				</tbody>
			</table>
		</div>
	</section>

	<section aria-labelledby="sets">
		<h2 id="sets">Bushing sets</h2>
		<p class="prose">Length is how far the bushing sticks out below the router base. Your template must be thicker than that, or the bushing will scrape the work.</p>
		{#each bushingSets as s (s.id)}
			<h3>{s.name}</h3>
			<p class="prose small">{shortSystem[s.system]}. {s.note}</p>
			<div class="scroll">
				<table>
					<thead><tr><th scope="col">Part</th><th scope="col">Outside</th><th scope="col">Inside</th><th scope="col">Length</th></tr></thead>
					<tbody>
						{#each s.bushings as b (b.id)}
							<tr><td>{b.part}</td><td>{size(b.od)}</td><td>{size(b.id_)}</td><td>{size(b.length)}</td></tr>
						{/each}
					</tbody>
				</table>
			</div>
		{/each}
		<p class="prose small">Inlay kits pair a 5/16 in bushing with a removable 9/16 in collar and a 1/8 in bit. The collar adds the bit's width twice, so one template cuts both the pocket and the insert. Bushwhack does the same job with two templates from one drawing: see <a href="/inlays">inlay templates</a>.</p>
	</section>

	<section aria-labelledby="bits">
		<h2 id="bits">Bits</h2>
		<p class="prose">Solid carbide spiral bits on a 1/4 in shank are the usual choice for template work. Upcut clears chips from deep cuts; downcut leaves a cleaner top edge on veneer and inlays. The cutting length must reach through the template into the work.</p>
		<div class="scroll">
			<table>
				<thead><tr><th scope="col">Bit</th><th scope="col">Diameter</th><th scope="col">Type</th><th scope="col">Cut length</th><th scope="col">Overall</th></tr></thead>
				<tbody>
					{#each bits as b (b.id)}
						<tr><th scope="row">{b.name}</th><td>{size(b.dia)}</td><td>{b.kind}</td><td>{size(b.cut)}</td><td>{size(b.overall)}</td></tr>
					{/each}
				</tbody>
			</table>
		</div>
	</section>

	<p><a class="btn" href="/make">Make a template</a></p>
	<p>See also the <a href="/guide-bushing-offset">offset calculator</a> and the <a href="/guide">guide</a>.</p>
</article>

<style>
	.page {
		padding-block: var(--s6) var(--s8);
	}
	.lead {
		font-size: var(--step-1);
		color: var(--text-soft);
	}
	section {
		margin-top: var(--s6);
	}
	h2 {
		font-size: var(--step-3);
	}
	.small {
		font-size: var(--step--1);
		color: var(--text-soft);
	}
	.scroll {
		overflow-x: auto;
		margin-bottom: var(--s3);
	}
	table {
		border-collapse: collapse;
		min-width: 24rem;
	}
	th,
	td {
		text-align: left;
		vertical-align: top;
		padding: 0.4rem 1.2rem 0.4rem 0;
		border-bottom: 1px solid var(--rule);
	}
	thead th {
		font-weight: 700;
		white-space: nowrap;
	}
	tbody th {
		font-weight: 600;
		white-space: nowrap;
	}
	.note-cell {
		min-width: 18rem;
		font-size: var(--step--1);
	}
</style>
