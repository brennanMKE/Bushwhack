<script lang="ts">
	import Seo from '#lib/components/Seo.svelte';
</script>

<Seo
	title="Router inlay templates from one drawing · Bushwhack"
	description="Make matched router inlay templates with a guide bushing: cut the pocket in hole mode and the insert in piece mode from the same SVG, so they fit."
	path="/inlays"
/>

<article class="wrap page">
	<h1>Router inlay templates</h1>
	<p class="prose lead">An inlay needs two cuts that fit each other: a pocket in the base board and an insert that drops into it. Bushwhack makes both templates from one drawing.</p>

	<section aria-labelledby="two">
		<h2 id="two">Two templates, one drawing</h2>
		<div class="prose">
			<ol>
				<li>Draw the inlay shape once and save it as an SVG.</li>
				<li>On the <a href="/make">workbench</a>, choose <strong>Cut a hole</strong> and download the template. This is the <strong>pocket</strong>: the hole in the base matches the drawing.</li>
				<li>Switch to <strong>Cut out a piece</strong>, keep the same size, bushing and bit, and download again. This is the <strong>insert</strong>: the part that falls out matches the drawing.</li>
			</ol>
			<p>With a 5/16 in bushing and a 1/8 in bit, the pocket template grows by 3/32 in and the insert template by 7/32 in. The <a href="/guide-bushing-offset">offset calculator</a> works out any other pair.</p>
			<p>This does the same job as a brass inlay kit, where a removable collar adds one bit diameter to the bushing. With two templates you need only the bushing, and the insert can be any shape you can draw.</p>
		</div>
	</section>

	<section aria-labelledby="corners">
		<h2 id="corners">Corners decide the fit</h2>
		<div class="prose">
			<p>A round bit can't cut a sharp inside corner. In the pocket, every point of the drawing (like the tips of a star) comes out rounded to the bit's radius. On the insert those same tips stay sharp, so they won't drop in.</p>
			<p>Round the outward points of your drawing to at least the bit's radius before you make either template: 1/16 in for a 1/8 in bit. Then pocket and insert match. The drawing's inward corners take care of themselves: sharp in the pocket, rounded on the insert, which leaves a tiny gap the glue fills.</p>
			<p>Small bits give crisper inlays. Use the smallest bit your bushing and stock allow.</p>
		</div>
	</section>

	<section aria-labelledby="cut">
		<h2 id="cut">Cutting and fitting</h2>
		<ul class="prose">
			<li>Rout the pocket about 1/8 in deep, in shallow passes.</li>
			<li>Cut the insert from stock a little thicker than the pocket is deep, or rout the insert's outline deeper than the pocket and resaw or plane it off.</li>
			<li>Test on scrap first. If the insert is tight, a strip of tape on the bushing makes the next insert slightly smaller.</li>
			<li>Glue it in proud and plane or sand flush.</li>
			<li>Keep shapes simple at small sizes. Bushwhack flags details narrower than the bit and strips of template under 1/4 in.</li>
		</ul>
	</section>

	<p><a class="btn" href="/make">Make inlay templates</a></p>
	<p>See also the <a href="/guide">guide</a> and <a href="/signs">sign and lettering templates</a>.</p>
</article>

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
	h2 {
		font-size: var(--step-3);
	}
	li {
		margin-bottom: var(--s1);
	}
</style>
