<script lang="ts">
	import { onMount } from 'svelte';
	import HeroDemo from '#lib/components/HeroDemo.svelte';
	import PatternTile from '#lib/components/PatternTile.svelte';
	import Seo from '#lib/components/Seo.svelte';
	import OffsetTable from '#lib/components/OffsetTable.svelte';
	import { SITE } from '#lib/site.ts';
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

<Seo
	title="Bushwhack: router templates that cut where you drew"
	description="Free router template maker. Upload an SVG, enter your guide bushing and bit, and get a true-size template offset so the cut lands on your line."
	path="/"
/>

<svelte:head>
	{@html `<script type="application/ld+json">${JSON.stringify({
		'@context': 'https://schema.org',
		'@type': 'WebApplication',
		name: 'Bushwhack',
		url: SITE + '/',
		description: 'Upload an SVG, enter your guide bushing and bit, and get a router template offset so the cut lands on your line.',
		applicationCategory: 'DesignApplication',
		operatingSystem: 'Any',
		offers: { '@type': 'Offer', price: '0', priceCurrency: 'USD' },
		author: { '@type': 'Person', name: 'Brennan Stehling', url: 'https://brennan.sstools.co/' }
	})}</script>`}
</svelte:head>

<section class="wrap hero">
	<div class="copy">
		<h1>Router templates that cut where you drew.</h1>
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

<section class="wrap learn" aria-labelledby="offset">
	<h2 id="offset">What is a guide bushing offset?</h2>
	<div class="prose">
		<p>A guide bushing is a metal collar in the router base. Its outside wall rides along the template while the bit spins in its centre, so the bit always cuts a fixed distance away from the template edge. That distance is the <strong>offset</strong>. Make a template at the exact size of your design and every opening comes out too small.</p>
		<p>Bushwhack grows every opening in your drawing by the offset, so the cut lands on your line.</p>
		<h3>The formula</h3>
		<p class="formula"><code>offset = (bushing OD − bit diameter) ÷ 2</code></p>
		<p><strong>Worked example:</strong> a 5/16 in bushing with a 1/8 in bit gives (0.3125 − 0.125) ÷ 2 = <strong>3/32 in (0.09375 in)</strong>. Every opening in the template is 3/32 in larger on each side than the hole you want.</p>
		<h3>Holes and pieces</h3>
		<p>That formula is for a <strong>hole</strong>: a pocket, recess or cut-out where the opening in your board should match the drawing. When the part that <strong>falls out</strong> should match the drawing, as for an inlay or a letter, the bit cuts on the far side of the line, so the offset is (bushing OD + bit diameter) ÷ 2. With the same setup, 7/32 in. In both cases the template opening is larger than the drawing.</p>
	</div>
	<OffsetTable caption="Template offsets for common bushing and bit pairs" />
	<p><a href="/guide-bushing-offset">Offset calculator for any bushing and bit</a></p>
</section>

<section class="wrap learn" aria-labelledby="faq">
	<h2 id="faq">Questions</h2>
	<div class="prose faq">
		<h3>What files does Bushwhack accept?</h3>
		<p>SVG, from Inkscape, Illustrator, Affinity, Figma or any CAD program. Convert text to paths first. You get two SVGs back: a true-size template for a laser, CNC or scroll saw, and a printable reference showing where the router will cut.</p>
		<h3>What should I make the template from?</h3>
		<p>1/4 in hardboard, MDF or acrylic. It must be thicker than the bushing collar is tall, or the collar will scrape your work. Leave about 2 in of border around the openings for clamping.</p>
		<h3>How sharp can an inside corner be?</h3>
		<p>The inside corners of a hole come out rounded to the bit's radius, because the bit is round. The template's own corners can't be tighter than the bushing's radius, so Bushwhack rounds them for you. For crisper corners use a smaller bit, or square them with a chisel.</p>
		<h3>Why does it warn about thin bridges?</h3>
		<p>Growing the openings thins the strips of template between them. Bushwhack flags any strip under 1/4 in, because it can snap or flex under the bushing. Make the design bigger or space the shapes apart.</p>
		<h3>Is it free? Are my files kept?</h3>
		<p>Free, with no account. Uploads are processed in memory and never stored.</p>
		<h3>Where do I learn more?</h3>
		<p>The <a href="/guide">guide</a> covers making and using templates. There are also pages on <a href="/inlays">inlay templates</a> and <a href="/signs">sign and lettering templates</a>, and free <a href="/patterns">starter patterns</a>.</p>
	</div>
</section>

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
	.learn {
		padding-block: var(--s4) var(--s6);
		border-top: 1px solid var(--rule);
	}
	.formula code {
		font-size: var(--step-1);
	}
	.faq h3:first-child {
		margin-top: 0;
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
