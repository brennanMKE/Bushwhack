<script lang="ts">
	import Seo from '#lib/components/Seo.svelte';
</script>

<Seo
	title="How to make a router template for a guide bushing · Bushwhack"
	description="How guide bushing offsets work, why corners round off, holes versus pieces, and how to make and rout with a template that lasts."
	path="/guide"
/>

<article class="wrap guide">
	<h1>Guide</h1>
	<p class="prose lead">How guide bushing offsets work, why corners round off, and how to make and use a template that lasts.</p>

	<section aria-labelledby="g-offset">
		<h2 id="g-offset">Why the template is bigger than your drawing</h2>
		<div class="split">
			<div class="prose">
				<p>A guide bushing is a metal collar screwed into the router base. The bit spins in its centre and the collar's outside wall rides along the template. So the bit never touches the template edge. It always runs a fixed distance away from it.</p>
				<p>Draw the template at the exact size of your design and every opening comes out too small, by the gap between the bushing wall and the bit on each side. Small details shrink to nothing. Bushwhack fixes that with offset geometry: it grows every opening outward by exactly that gap.</p>
				<p>The bit cuts a band. The near edge of the band is the bushing radius minus the bit radius from the template. The far edge is the bushing radius plus the bit radius.</p>
			</div>
			<figure class="diagram">
				<svg viewBox="0 0 420 300" role="img" aria-label="Top view: the bushing touches the template edge, and the bit spins at the bushing centre">
					<defs>
						<pattern id="g-hatch" width="8" height="8" patternUnits="userSpaceOnUse" patternTransform="rotate(45)">
							<path d="M0 0V8" stroke="var(--hardboard)" stroke-width="3" />
						</pattern>
					</defs>
					<rect x="0" y="20" width="110" height="250" fill="url(#g-hatch)" />
					<line x1="110" y1="20" x2="110" y2="270" stroke="var(--hardboard)" stroke-width="3" />
					<text x="55" y="292" text-anchor="middle">template</text>
					<circle cx="200" cy="145" r="90" fill="var(--brass)" fill-opacity="0.15" stroke="var(--brass)" stroke-width="4" />
					<text x="200" y="44" text-anchor="middle">bushing</text>
					<circle cx="200" cy="145" r="36" fill="currentColor" />
					<text x="200" y="150" text-anchor="middle" fill="var(--bg)" font-weight="700">bit</text>
					<line x1="164" y1="20" x2="164" y2="270" stroke="var(--cut)" stroke-width="2.5" />
					<line x1="236" y1="20" x2="236" y2="270" stroke="var(--cut)" stroke-width="2.5" />
					<text x="164" y="14" text-anchor="middle" class="sm">hole wall</text>
					<text x="236" y="14" text-anchor="middle" class="sm">piece edge</text>
					<g stroke="currentColor" stroke-width="1.5" fill="none">
						<path d="M110 215H164M110 209v12M164 209v12" />
						<path d="M110 250H236M236 244v12" />
					</g>
					<text x="137" y="237" text-anchor="middle" class="sm">hole offset</text>
					<text x="310" y="254" text-anchor="middle" class="sm">piece offset</text>
				</svg>
				<figcaption>Top view. Red lines are the two edges of the band the bit cuts.</figcaption>
			</figure>
		</div>
	</section>

	<section aria-labelledby="g-modes">
		<h2 id="g-modes">Holes and pieces</h2>
		<div class="prose">
			<p><strong>Cut a hole</strong> when the opening in your board should match the drawing: pumpkin faces, pockets, recesses, cut-outs. The template grows by the bushing radius minus the bit radius. With a 5/16 in bushing and a 1/8 in bit that's 3/32 in.</p>
			<p><strong>Cut out a piece</strong> when the part that falls out should match the drawing, as with inlays or letters. The template grows by the bushing radius plus the bit radius. Same setup: 7/32 in.</p>
			<p>For a matched inlay you'll usually cut the pocket in hole mode and the insert in piece mode, from the same drawing.</p>
		</div>
	</section>

	<section aria-labelledby="g-corners">
		<h2 id="g-corners">Corners and small details</h2>
		<div class="prose">
			<p><strong>Inside corners round off.</strong> A spinning bit is round, so a sharp inside corner of a hole always comes out with a radius equal to the bit's radius. Use a smaller bit if that matters, or square the corners with a chisel.</p>
			<p><strong>Outside corners of a piece stay sharp.</strong> The bit's far edge wraps around them.</p>
			<p><strong>The template's corners are rounded on purpose.</strong> The bushing can't get into a sharper corner than its own radius anyway, so rounding the template costs nothing and avoids fragile spikes.</p>
			<p><strong>Details narrower than the bit can't be cut.</strong> Bushwhack shows them with a red hatch and reports how much of each shape is lost. Make the design bigger or use a smaller bit.</p>
			<p><strong>Openings can't be too close together.</strong> When openings grow, the strips of template between them get thinner. Bushwhack flags any bridge under 1/4 in, and draws a bushing at true size so you can see the problem. When openings would touch, they merge into one.</p>
		</div>
	</section>

	<section aria-labelledby="g-make">
		<h2 id="g-make">Making the template</h2>
		<div class="prose">
			<ul>
				<li>Use 1/4 in hardboard, MDF or acrylic. It must be thicker than the bushing's collar is tall, or the collar will scrape your work.</li>
				<li>Import the template SVG at 100%. Its size is set in inches, so never scale it to fit. Measure one opening with calipers before you rout.</li>
				<li>Cut the openings on the line. On a laser, the kerf is small enough to ignore. With a scroll saw, cut just outside the line and sand to it.</li>
				<li>Leave a border of at least 2 in around the openings for clamping and for the router base to rest on.</li>
				<li>A shape with a hole in its middle (like an O) leaves a loose island in the template. Mount the template on a backer or add small bridges.</li>
			</ul>
		</div>
	</section>

	<section aria-labelledby="g-rout">
		<h2 id="g-rout">Routing safely</h2>
		<div class="prose">
			<ul>
				<li>Wear eye and hearing protection, and use dust collection.</li>
				<li>Clamp the template and the work so neither can shift. Double-sided tape works for small pieces.</li>
				<li>Check the bit clears the inside of the bushing before you switch on.</li>
				<li>Plunge in several shallow passes, about 1/8 in each, rather than one deep one.</li>
				<li>Keep the bushing pressed against the template. Inside an opening, move the router clockwise.</li>
				<li>Let the bit stop before you lift the router out of the template.</li>
			</ul>
		</div>
	</section>

	<section aria-labelledby="g-more">
		<h2 id="g-more">More</h2>
		<ul class="prose">
			<li><a href="/guide-bushing-offset">Guide bushing offset calculator</a>, with a table of common pairs.</li>
			<li><a href="/gear">Routers, bushings and bits</a>: sizes, and which bushings fit which router.</li>
			<li><a href="/inlays">Inlay templates</a>: a pocket and an insert from one drawing.</li>
			<li><a href="/signs">Sign and lettering templates</a>.</li>
			<li><a href="/patterns">Starter patterns</a> for holidays and special days.</li>
		</ul>
	</section>

	<p><a class="btn" href="/make">Make a template</a></p>
</article>

<style>
	.guide {
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
	.split {
		display: grid;
		grid-template-columns: minmax(0, 1fr) minmax(0, 26rem);
		gap: var(--s4);
		align-items: start;
	}
	.diagram {
		margin: 0;
	}
	.diagram svg {
		width: 100%;
		height: auto;
		font-family: var(--font-text);
		font-size: 15px;
		fill: currentColor;
	}
	.diagram .sm {
		font-size: 13px;
	}
	figcaption {
		font-size: var(--step--1);
		color: var(--text-soft);
	}
	li {
		margin-bottom: var(--s1);
	}
	@media (max-width: 800px) {
		.split {
			grid-template-columns: 1fr;
		}
	}
</style>
