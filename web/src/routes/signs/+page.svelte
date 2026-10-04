<script lang="ts">
	import Seo from '#lib/components/Seo.svelte';
</script>

<Seo
	title="Router sign and lettering templates · Bushwhack"
	description="Make router templates for signs and lettering with a guide bushing: recessed letters in hole mode, cut-out or inlaid letters in piece mode, with font and size tips."
	path="/signs"
/>

<article class="wrap page">
	<h1>Sign and lettering templates</h1>
	<p class="prose lead">Lay out your sign in any drawing program, save it as SVG, and Bushwhack makes a template that routes the letters at the size you drew them.</p>

	<section aria-labelledby="prep">
		<h2 id="prep">Preparing the artwork</h2>
		<ul class="prose">
			<li><strong>Convert text to paths</strong> (outlines) before you save. Bushwhack reads shapes, not fonts. In Inkscape it's Path › Object to Path; in Illustrator, Type › Create Outlines.</li>
			<li><strong>Pick a sturdy font.</strong> Bold sans-serifs and slab serifs rout well. Thin strokes and fine serifs narrower than the bit can't be cut.</li>
			<li><strong>Space the letters out.</strong> The template between letters gets thinner as the openings grow. Add letter spacing, or Bushwhack will flag bridges under 1/4 in or merge letters that touch.</li>
			<li><strong>Size it big enough.</strong> With a 1/8 in bit, letters about 2 in tall and up in a bold font usually cut cleanly. The workbench shows what's lost at your size.</li>
		</ul>
	</section>

	<section aria-labelledby="modes">
		<h2 id="modes">Recessed, cut-out or inlaid</h2>
		<div class="prose">
			<p><strong>Recessed letters</strong> are pockets routed into the sign. Use <strong>Cut a hole</strong>, so each letter-shaped hole matches your artwork. Paint the pockets before finishing for contrast.</p>
			<p><strong>Cut-out letters</strong> are routed all the way through and mounted on a backer. Use <strong>Cut out a piece</strong>, so the letters that fall out match your artwork.</p>
			<p><strong>Inlaid letters</strong> use both: a pocket in the sign and letters cut from a contrasting wood. See <a href="/inlays">inlay templates</a>.</p>
		</div>
	</section>

	<section aria-labelledby="counters">
		<h2 id="counters">Letters with holes</h2>
		<div class="prose">
			<p>The middle of an O, A, B, D, P, Q or R (the counter) becomes a loose island in the template. Mount the template on a clear backer, add thin bridges to the letters in your artwork (a stencil font has them built in), or cut those letters in two passes. Bushwhack points out every island it finds.</p>
		</div>
	</section>

	<p><a class="btn" href="/make">Make a sign template</a></p>
	<p>See also the <a href="/guide">guide</a> and the <a href="/guide-bushing-offset">offset calculator</a>.</p>
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
