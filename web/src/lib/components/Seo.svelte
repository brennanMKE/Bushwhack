<script lang="ts">
	// Per-page title, description and canonical URL, plus the matching Open
	// Graph and Twitter tags. Site-wide tags live in app.html.
	import { canonical } from '#lib/site.ts';

	let { title, description, path }: { title: string; description: string; path: string } = $props();
	const url = $derived(canonical(path));
</script>

<svelte:head>
	<title>{title}</title>
	<meta name="description" content={description} />
	<link rel="canonical" href={url} />
	<meta property="og:title" content={title} />
	<meta property="og:description" content={description} />
	<meta property="og:url" content={url} />
	<meta name="twitter:title" content={title} />
	<meta name="twitter:description" content={description} />
</svelte:head>
