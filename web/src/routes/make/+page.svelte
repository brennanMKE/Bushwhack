<script lang="ts">
	import { onMount, untrack } from 'svelte';
	import { page } from '$app/state';
	import { goto } from '$app/navigation';
	import MatPreview from '#lib/components/MatPreview.svelte';
	import MeasureInput from '#lib/components/MeasureInput.svelte';
	import GearPicker from '#lib/components/GearPicker.svelte';
	import { bitFit, innerFor, standardInner, same } from '#lib/gear.ts';
	import ModeToggle from '#lib/components/ModeToggle.svelte';
	import PatternTile from '#lib/components/PatternTile.svelte';
	import {
		ApiError,
		loadPatterns,
		processSvg,
		shapeLabel,
		type Note,
		type Occasion,
		type Pattern,
		type ProcessResult
	} from '#lib/api.ts';
	import { parseLength, fraction } from '#lib/units.ts';
	import { loadSettings, saveSettings } from '#lib/settings.ts';

	type Source = { kind: 'file'; file: File; name: string } | { kind: 'pattern'; pattern: Pattern };

	const saved = loadSettings();
	let bushing = $state(saved.bushing);
	let bit = $state(saved.bit);
	let mode = $state<'hole' | 'piece'>(saved.mode);
	let fit = $state<'canvas' | 'artwork'>('canvas');
	let size = $state('6');
	let minBridge = $state(saved.minBridge);
	let inner = $state(saved.inner);
	let router = $state(saved.router);
	let bushingPick = $state(saved.bushingPick);
	let bitPick = $state(saved.bitPick);
	const innerNow = $derived(innerFor(bushing, inner, bushingPick).value);
	const fitCheck = $derived(bitFit(bushing, innerNow, bit));
	let useDocument = $state(false);
	let scale = $state('');

	let source = $state<Source | null>(null);
	let result = $state<ProcessResult | null>(null);
	let error = $state('');
	let busy = $state(false);
	let dragOver = $state(false);
	let highlight = $state<number | null>(null);
	let show = $state({ drawing: true, template: true, cut: true });
	let pulse = $state({ bridge: -1, key: 0 });
	let toast = $state('');
	let seasonal = $state<Occasion | null>(null);
	let fileInput = $state<HTMLInputElement | undefined>();
	let previewEl = $state<HTMLElement | undefined>();

	// Values that arrived in a shared link aren't saved as this browser's
	// defaults until the viewer changes one of them.
	let fromLink = '';
	const settingsKey = () => JSON.stringify([bushing, inner, bit, mode, minBridge]);
	$effect(() => {
		const key = settingsKey();
		const picks = { router, bushingPick, bitPick };
		if (key !== fromLink) saveSettings({ bushing, bit, mode, minBridge, inner, ...picks });
	});

	const valid = $derived(
		fitCheck.ok &&
		[bushing, bit, minBridge, ...(useDocument || scale.trim() ? [] : [size])].every((v) => {
			const p = parseLength(v);
			return p !== null && p.inches > 0;
		})
	);
	const sizeIn = $derived(parseLength(size)?.inches ?? 0);
	const offsetText = $derived.by(() => {
		const b = parseLength(bushing)?.inches;
		const d = parseLength(bit)?.inches;
		if (!b || !d || d >= b) return '';
		const o = mode === 'hole' ? b / 2 - d / 2 : b / 2 + d / 2;
		return `${fraction(o)} (${o.toFixed(4)} in)`;
	});

	// Settings from the URL (?bushing=5/16&bit=1/8&mode=hole&size=5) win over
	// saved ones, so a shared link reproduces the same job. Re-applied when
	// back/forward lands on a different URL.
	const lengthParam = (q: { get(k: string): string | null }, k: string) => {
		const v = q.get(k)?.trim();
		return v && parseLength(v) ? v : null;
	};
	$effect(() => {
		const q = page.url.searchParams;
		untrack(() => {
			const b = lengthParam(q, 'bushing');
			if (b) {
				bushing = b;
				inner = lengthParam(q, 'inner') ?? '';
			}
			bit = lengthParam(q, 'bit') ?? bit;
			const m = q.get('mode');
			if (m === 'hole' || m === 'piece') mode = m;
			size = lengthParam(q, 'size') ?? size;
			if (q.has('bushing') || q.has('bit') || q.has('mode')) fromLink = settingsKey();
		});
	});

	// Keep the URL in step with the settings, without adding history entries.
	let urlTimer: ReturnType<typeof setTimeout> | undefined;
	$effect(() => {
		const want = {
			bushing,
			// The inside diameter only when it isn't the standard one for the size.
			inner: innerNow && !same(innerNow, standardInner(bushing) ?? '') ? innerNow : '',
			bit,
			mode,
			size: useDocument || scale.trim() ? '' : size,
			pattern: source?.kind === 'pattern' ? source.pattern.slug : source?.kind === 'file' ? '' : null
		};
		clearTimeout(urlTimer);
		urlTimer = setTimeout(() => untrack(() => syncURL(want)), 250);
		return () => clearTimeout(urlTimer);
	});

	function syncURL(want: Record<string, string | null>) {
		const url = new URL(location.href);
		if (url.pathname !== '/make') return;
		const q = new URLSearchParams();
		// pattern null: nothing chosen yet, keep whatever the link asked for.
		const pattern = want.pattern ?? url.searchParams.get('pattern');
		if (pattern) q.set('pattern', pattern);
		for (const k of ['bushing', 'inner', 'bit', 'mode', 'size']) if (want[k]?.trim()) q.set(k, want[k]!.trim());
		// "/" is legal in a query; keep 5/16 readable.
		const search = q.toString().replace(/%2F/gi, '/');
		if (search === url.search.slice(1)) return;
		goto(`/make${search ? '?' + search : ''}`, { replace: true, shallow: true });
	}

	// Pattern from ?pattern=slug, also when navigating between patterns here.
	$effect(() => {
		const slug = page.url.searchParams.get('pattern');
		if (!slug) return;
		const urlSize = lengthParam(page.url.searchParams, 'size');
		untrack(() => {
			if (source?.kind === 'pattern' && source.pattern.slug === slug) return;
			loadPatterns()
				.then((all) => {
					const p = all.flatMap((o) => o.patterns).find((x) => x.slug === slug);
					if (!p) {
						error = "That pattern doesn't exist. Pick another or drop in your own SVG.";
						return;
					}
					size = urlSize ?? String(p.recommendedSizeIn);
					fit = p.fit === 'artwork' ? 'artwork' : 'canvas';
					useDocument = false;
					scale = '';
					source = { kind: 'pattern', pattern: p };
				})
				.catch(() => (error = 'Could not load that pattern. Check your connection and try again.'));
		});
	});

	onMount(async () => {
		try {
			const all = await loadPatterns();
			seasonal = all.find((o) => o.upcoming && o.patterns.length) ?? null;
		} catch {
			seasonal = null;
		}
	});

	// Re-run 300 ms after the last change. Old result stays visible meanwhile.
	let timer: ReturnType<typeof setTimeout> | undefined;
	let inflight: AbortController | null = null;
	$effect(() => {
		const job = { source, bushing, innerNow, bit, mode, fit, size, minBridge, useDocument, scale, valid };
		if (!job.source || !job.valid) return;
		clearTimeout(timer);
		timer = setTimeout(() => untrack(() => run()), 300);
		return () => clearTimeout(timer);
	});

	async function run() {
		if (!source) return;
		inflight?.abort();
		const ctl = new AbortController();
		inflight = ctl;
		const form = new FormData();
		if (source.kind === 'file') form.set('file', source.file);
		else form.set('pattern', source.pattern.slug);
		form.set('bushing', bushing);
		if (innerNow) form.set('bushingId', innerNow);
		form.set('bit', bit);
		form.set('mode', mode);
		form.set('fit', useDocument ? 'document' : fit);
		form.set('size', size);
		form.set('minBridge', minBridge);
		if (scale.trim()) form.set('scale', scale.trim());
		busy = true;
		try {
			const r = await processSvg(form, ctl.signal);
			if (inflight !== ctl) return;
			result = r;
			error = '';
			if (highlight != null && !r.shapes.some((s) => s.index === highlight)) highlight = null;
		} catch (e) {
			if ((e as Error).name === 'AbortError') return;
			error = e instanceof ApiError ? e.message : 'Could not reach Bushwhack. Check your connection and try again.';
		} finally {
			if (inflight === ctl) busy = false;
		}
	}

	function choose(files: FileList | null | undefined) {
		const f = files?.[0];
		if (!f) return;
		if (f.size > 2 * 1024 * 1024) {
			error = 'That file is over 2 MB. Simplify the drawing or save it as a plain SVG.';
			return;
		}
		if (!/\.svg$/i.test(f.name) && f.type !== 'image/svg+xml') {
			error = "That doesn't look like an SVG. Export your design as .svg and try again.";
			return;
		}
		error = '';
		source = { kind: 'file', file: f, name: f.name };
	}

	function onDrop(e: DragEvent) {
		e.preventDefault();
		dragOver = false;
		choose(e.dataTransfer?.files);
	}

	const baseName = $derived(
		source?.kind === 'file'
			? source.name.replace(/\.svg$/i, '').replace(/[^\w.-]+/g, '-') || 'drawing'
			: source?.kind === 'pattern'
				? source.pattern.slug
				: 'drawing'
	);

	function download(kind: 'template' | 'preview') {
		if (!result) return;
		const svg = kind === 'template' ? result.templateSvg : result.previewSvg;
		const url = URL.createObjectURL(new Blob([svg], { type: 'image/svg+xml' }));
		const a = document.createElement('a');
		a.href = url;
		a.download = `${baseName}-${kind}.svg`;
		document.body.append(a);
		a.click();
		a.remove();
		setTimeout(() => URL.revokeObjectURL(url), 1000);
		toast = kind === 'template' ? 'Template downloaded' : 'Preview downloaded';
		setTimeout(() => (toast = ''), 2600);
	}

	function focusNote(n: Note) {
		if (n.bridge >= 0) {
			previewEl?.scrollIntoView({ behavior: 'smooth', block: 'nearest' });
			pulse = { bridge: n.bridge, key: pulse.key + 1 };
		} else if (n.shapes?.length) {
			highlight = n.shapes[0];
			previewEl?.scrollIntoView({ behavior: 'smooth', block: 'nearest' });
		}
	}

	const summary = $derived(
		result
			? `Preview on a cutting mat. Page ${result.page.widthIn.toFixed(3)} by ${result.page.heightIn.toFixed(3)} inches, ${result.shapes.length} ${result.shapes.length === 1 ? 'opening' : 'openings'}, ${result.notes.length ? `${result.notes.length} ${result.notes.length === 1 ? 'warning' : 'warnings'}` : 'no warnings'}.`
			: ''
	);
	const sourceName = $derived(
		source?.kind === 'file' ? source.name : source?.kind === 'pattern' ? `${source.pattern.name} pattern` : ''
	);
	const dims = (w: number, h: number) => `${w.toFixed(2)} × ${h.toFixed(2)}`;
</script>

<svelte:head><title>Make a template · Bushwhack</title></svelte:head>

<div class="wrap bench">
	<section class="controls" aria-label="Settings">
		<h1 class="visually-hidden">Make a template</h1>
		<div class="group">
			<h2 class="label">Your file</h2>
			<label
				class="drop"
				class:over={dragOver}
				ondragover={(e) => {
					e.preventDefault();
					dragOver = true;
				}}
				ondragleave={() => (dragOver = false)}
				ondrop={onDrop}
			>
				<input bind:this={fileInput} type="file" accept=".svg,image/svg+xml" onchange={(e) => choose(e.currentTarget.files)} />
				{#if sourceName}
					<strong class="fname">{sourceName}</strong>
					<span>Drop another SVG or click to replace it</span>
				{:else}
					<strong>Drop an SVG here</strong>
					<span>or click to choose one</span>
				{/if}
			</label>
		</div>

		<div class="group">
			<GearPicker bind:bushing bind:inner bind:bit bind:router bind:bushingPick bind:bitPick />
			{#if fitCheck.error}
				<p class="note" role="alert">{fitCheck.error}</p>
			{:else if fitCheck.warn}
				<p class="note">{fitCheck.warn}</p>
			{/if}
			{#if offsetText && fitCheck.ok}<p class="offset">Template grows by <strong>{offsetText}</strong></p>{/if}
		</div>

		<div class="group">
			<ModeToggle bind:mode />
		</div>

		<div class="group">
			<h2 class="label">Size</h2>
			{#if !useDocument && !scale.trim()}
				<MeasureInput id="size" label="Longest side" bind:value={size} />
				<div class="fit pill" role="radiogroup" aria-label="What the longest side measures">
					<label class:on={fit === 'canvas'}><input type="radio" name="fit" value="canvas" bind:group={fit} />Whole page</label>
					<label class:on={fit === 'artwork'}><input type="radio" name="fit" value="artwork" bind:group={fit} />Just the shapes</label>
				</div>
				{#if sizeIn > 0}<p class="help">Longest side of the {fit === 'canvas' ? 'page' : 'shapes'} becomes {fraction(sizeIn)}.</p>{/if}
			{:else}
				<p class="help">{useDocument ? "Using the size saved in the SVG." : 'Using your scale.'}</p>
			{/if}
		</div>

		<details class="group more">
			<summary>More settings</summary>
			<MeasureInput id="minBridge" label="Minimum bridge" bind:value={minBridge} hint="thinner template wood is flagged" />
			<label class="check"><input type="checkbox" bind:checked={useDocument} /> Use the size saved in the SVG (in, mm or cm)</label>
			<label class="plain" for="scale">Scale (inches per SVG unit)</label>
			<input id="scale" class="plain-input" bind:value={scale} placeholder="automatic" inputmode="decimal" autocomplete="off" />
		</details>

		{#if valid}
			<p class="help bookmark">
				<strong>Bookmark this page</strong> to keep these settings. The address holds your bushing, bit, mode and size, and bookmarks sync to your other devices.
			</p>
		{/if}
	</section>

	<section class="preview" aria-label="Preview" bind:this={previewEl}>
		{#if result}
			<div class="canvas">
				<MatPreview {result} {busy} {highlight} {show} {pulse} label={summary} />
			</div>
			<div class="legend" role="group" aria-label="Layers">
				<label><input type="checkbox" bind:checked={show.drawing} /><i class="k drawing"></i>Drawing</label>
				<label><input type="checkbox" bind:checked={show.template} /><i class="k template"></i>Template</label>
				<label><input type="checkbox" bind:checked={show.cut} /><i class="k cut"></i>Cut</label>
				<span class="k-note"><i class="k lost"></i>Bit can't reach</span>
			</div>
		{:else}
			<div class="canvas empty" class:busy>
				<div class="empty-inner">
					<p class="empty-title">Drop an SVG here, or start with a pattern.</p>
					<button type="button" class="btn" onclick={() => fileInput?.click()}>Choose an SVG</button>
					{#if seasonal}
						<div class="starter">
							{#each seasonal.patterns.slice(0, 3) as p (p.slug)}
								<PatternTile pattern={p} href="/make?pattern={p.slug}" />
							{/each}
						</div>
					{/if}
				</div>
			</div>
		{/if}
		{#if error}
			<div class="note err" role="alert">
				<svg viewBox="0 0 20 20" aria-hidden="true"><circle cx="10" cy="10" r="8.5" fill="none" stroke="currentColor" stroke-width="1.8" /><path d="M10 5.5v5.5M10 13.6v.9" stroke="currentColor" stroke-width="2" stroke-linecap="round" /></svg>
				<span>{error}</span>
			</div>
		{/if}
	</section>

	{#if result}
		<section class="results" aria-label="Results">
			<p class="readout"><span class="visually-hidden">Template page size </span>{result.page.widthIn.toFixed(3)} × {result.page.heightIn.toFixed(3)} in</p>
			<dl class="stats">
				<div><dt>Template grows</dt><dd>{fraction(result.offsetIn)}</dd></div>
				<div><dt>Thinnest bridge</dt><dd>{result.minBridgeIn == null ? 'n/a' : `${result.minBridgeIn.toFixed(3)} in`}</dd></div>
				<div><dt>Openings</dt><dd>{result.shapes.length}</dd></div>
			</dl>

			<div class="actions">
				<button type="button" class="btn" onclick={() => download('template')}>Download template</button>
				<button type="button" class="btn quiet" onclick={() => download('preview')}>Download preview</button>
				<p class="import">Import at 100%. Don't resize it, or the bushing math is off.</p>
			</div>

			<div class="notes" aria-live="polite">
				{#each result.notes as n, i (i + n.text)}
					{#if n.bridge >= 0 || n.shapes?.length}
						<button type="button" class="note" onclick={() => focusNote(n)}>
							<svg viewBox="0 0 20 20" aria-hidden="true"><circle cx="10" cy="10" r="8.5" fill="none" stroke="currentColor" stroke-width="1.8" /><path d="M10 5.5v5.5M10 13.6v.9" stroke="currentColor" stroke-width="2" stroke-linecap="round" /></svg>
							<span>{n.text} <span class="show">Show me</span></span>
						</button>
					{:else}
						<div class="note">
							<svg viewBox="0 0 20 20" aria-hidden="true"><circle cx="10" cy="10" r="8.5" fill="none" stroke="currentColor" stroke-width="1.8" /><path d="M10 5.5v5.5M10 13.6v.9" stroke="currentColor" stroke-width="2" stroke-linecap="round" /></svg>
							<span>{n.text}</span>
						</div>
					{/if}
				{/each}
			</div>

			<div class="table-wrap">
				<table>
					<thead>
						<tr><th scope="col">Shape</th><th scope="col" class="num">Size (in)</th><th scope="col" class="num">Template (in)</th><th scope="col" class="num">Lost</th></tr>
					</thead>
					<tbody>
						{#each result.shapes as s (s.index)}
							<tr
								tabindex="0"
								class:on={highlight === s.index}
								onmouseenter={() => (highlight = s.index)}
								onmouseleave={() => (highlight = null)}
								onfocus={() => (highlight = s.index)}
								onblur={() => (highlight = null)}
							>
								<th scope="row">{shapeLabel(s)}</th>
								<td class="num">{dims(s.widthIn, s.heightIn)}</td>
								<td class="num">{dims(s.templateWidthIn, s.templateHeightIn)}</td>
								<td class="num">{s.lostAreaPct.toFixed(1)}%</td>
							</tr>
						{/each}
					</tbody>
				</table>
			</div>
		</section>
	{/if}
</div>

<div class="toast" role="status" aria-live="polite">{#if toast}<span>{toast}</span>{/if}</div>

<style>
	.bench {
		display: grid;
		grid-template-columns: 320px minmax(0, 1fr);
		grid-template-areas:
			'controls preview'
			'controls results';
		grid-template-rows: auto 1fr;
		gap: var(--s3) var(--s4);
		padding-block: var(--s3) var(--s6);
		align-items: start;
	}
	.controls {
		grid-area: controls;
		display: grid;
		gap: var(--s3);
	}
	.preview {
		grid-area: preview;
		display: grid;
		gap: var(--s1);
	}
	.results {
		grid-area: results;
		display: grid;
		gap: var(--s3);
	}
	.group {
		display: grid;
		gap: var(--s2);
	}
	.label {
		font-family: var(--font-text);
		font-size: var(--step-0);
		font-weight: 700;
		margin: 0;
		line-height: 1.3;
	}
	.drop {
		position: relative;
		display: grid;
		justify-items: center;
		gap: 0.2rem;
		padding: var(--s3) var(--s2);
		border: 2px dashed var(--line);
		border-radius: var(--radius);
		background: var(--surface);
		text-align: center;
		cursor: pointer;
		overflow-wrap: anywhere;
	}
	.drop.over,
	.drop:hover {
		border-color: var(--brass);
	}
	.drop:has(input:focus-visible) {
		outline: 2px solid var(--focus);
		outline-offset: 2px;
	}
	.drop input {
		position: absolute;
		inset: 0;
		opacity: 0;
		cursor: pointer;
	}
	.drop span {
		font-size: var(--step--1);
		color: var(--text-soft);
	}
	.offset,
	.help {
		margin: 0;
		font-size: var(--step--1);
		color: var(--text-soft);
	}
	.bookmark {
		padding: 0.5rem 0.7rem;
		border-left: 4px solid var(--brass);
		background: color-mix(in srgb, var(--brass) 10%, transparent);
	}
	.pill {
		display: grid;
		grid-template-columns: 1fr 1fr;
		border: 1.5px solid var(--line);
		border-radius: 999px;
		padding: 3px;
		background: var(--surface);
	}
	.pill label {
		position: relative;
		text-align: center;
		padding: 0.4rem 0.5rem;
		border-radius: 999px;
		font-weight: 600;
		cursor: pointer;
	}
	.pill label.on {
		background: var(--brass);
		color: var(--graphite);
	}
	.pill label:has(input:focus-visible) {
		outline: 2px solid var(--focus);
		outline-offset: 2px;
	}
	.pill input {
		position: absolute;
		inset: 0;
		opacity: 0;
		margin: 0;
		cursor: pointer;
	}
	.more summary {
		font-weight: 700;
		cursor: pointer;
	}
	.check {
		display: flex;
		gap: 0.5rem;
		align-items: flex-start;
		font-size: var(--step--1);
	}
	.check input {
		margin-top: 0.25rem;
		accent-color: var(--brass);
	}
	.plain {
		font-weight: 700;
		margin-bottom: -0.6rem;
	}
	.plain-input {
		font: inherit;
		padding: 0.45rem 0.65rem;
		border: 1.5px solid var(--line);
		border-radius: var(--radius);
		background: var(--surface);
		color: var(--text);
	}

	.canvas {
		height: min(68vh, 640px);
		min-height: 320px;
	}
	.empty {
		display: grid;
		place-items: center;
		background: var(--mat);
		color: var(--shop-wall);
		padding: var(--s3);
		background-image:
			linear-gradient(var(--mat-grid) 1px, transparent 1px),
			linear-gradient(90deg, var(--mat-grid) 1px, transparent 1px);
		background-size: 64px 64px;
		background-position: -1px -1px;
	}
	.empty-inner {
		background: var(--mat);
		padding: var(--s3);
		display: grid;
		justify-items: start;
		gap: var(--s2);
		max-width: 30rem;
	}
	.empty-title {
		font-family: var(--font-display);
		font-weight: 700;
		font-size: var(--step-3);
		line-height: 1;
		margin: 0;
	}
	.starter {
		display: grid;
		grid-template-columns: repeat(3, 1fr);
		gap: var(--s2);
		width: 100%;
		margin-top: var(--s1);
	}
	.starter :global(a) {
		color: var(--shop-wall);
	}
	.legend {
		display: flex;
		flex-wrap: wrap;
		gap: var(--s1) var(--s3);
		font-size: var(--step--1);
		font-weight: 600;
	}
	.legend label,
	.k-note {
		display: inline-flex;
		align-items: center;
		gap: 0.4rem;
		cursor: pointer;
	}
	.k-note {
		cursor: default;
		color: var(--text-soft);
	}
	.legend input {
		accent-color: var(--brass);
	}
	.k {
		display: inline-block;
		width: 1.6rem;
		height: 0.9rem;
		background: var(--mat);
		border-radius: 1px;
		position: relative;
	}
	.k::after {
		content: '';
		position: absolute;
		left: 3px;
		right: 3px;
		top: 50%;
		margin-top: -1px;
		height: 0;
	}
	.k.drawing::after {
		border-top: 2px dashed var(--chalk-on-mat);
	}
	.k.cut::after {
		border-top: 2.5px solid var(--cut-on-mat);
	}
	.k.template {
		background: var(--hardboard);
		outline: 1px solid var(--brass);
		outline-offset: -1px;
	}
	.k.lost {
		background: repeating-linear-gradient(45deg, var(--cut-on-mat) 0 1.5px, var(--mat) 1.5px 5px);
	}
	.err {
		margin-top: var(--s1);
	}

	.readout {
		font-family: var(--font-display);
		font-weight: 800;
		font-size: clamp(2.6rem, 2rem + 2.5vw, 4rem);
		line-height: 1;
		margin: 0;
		letter-spacing: 0.01em;
	}
	.stats {
		display: flex;
		flex-wrap: wrap;
		gap: var(--s1) var(--s4);
		margin: 0;
	}
	.stats div {
		display: grid;
	}
	.stats dt {
		font-size: var(--step--1);
		color: var(--text-soft);
	}
	.stats dd {
		margin: 0;
		font-weight: 700;
		font-size: var(--step-1);
	}
	.actions {
		display: flex;
		flex-wrap: wrap;
		align-items: center;
		gap: var(--s1) var(--s2);
	}
	.import {
		flex-basis: 100%;
		margin: 0;
		font-weight: 600;
	}
	.notes {
		display: grid;
		gap: var(--s1);
	}
	.show {
		font-weight: 700;
		text-decoration: underline;
		text-decoration-color: var(--brass);
		text-underline-offset: 3px;
		white-space: nowrap;
	}
	.table-wrap {
		overflow-x: auto;
	}
	table {
		width: 100%;
		border-collapse: collapse;
		font-variant-numeric: tabular-nums;
	}
	th,
	td {
		text-align: left;
		padding: 0.45rem 0.6rem;
		border-bottom: 1px solid var(--rule);
	}
	thead th {
		font-size: var(--step--1);
		color: var(--text-soft);
		font-weight: 600;
	}
	.num {
		text-align: right;
		white-space: nowrap;
	}
	@media (max-width: 480px) {
		th,
		td {
			padding: 0.4rem 0.3rem;
			font-size: 0.92rem;
		}
	}
	tbody tr {
		cursor: default;
	}
	tbody tr.on,
	tbody tr:focus-visible {
		background: color-mix(in srgb, var(--brass) 16%, transparent);
	}
	tbody tr:focus-visible {
		outline: 2px solid var(--focus);
		outline-offset: -2px;
	}
	.toast {
		position: fixed;
		left: 50%;
		bottom: var(--s3);
		transform: translateX(-50%);
		z-index: 20;
	}
	.toast span {
		display: block;
		background: var(--graphite);
		color: var(--shop-wall);
		border-left: 4px solid var(--brass);
		padding: 0.6rem 1rem;
		font-weight: 600;
	}

	@media (max-width: 900px) {
		.bench {
			grid-template-columns: minmax(0, 1fr);
			grid-template-areas: 'preview' 'controls' 'results';
			grid-template-rows: auto;
		}
		.canvas {
			height: min(62vh, 520px);
		}
	}
</style>
