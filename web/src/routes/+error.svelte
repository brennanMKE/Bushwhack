<script lang="ts">
	import { page } from '$app/state';
</script>

<svelte:head><title>Not found · Bushwhack</title></svelte:head>

<div class="wrap err">
	<h1>{page.status === 404 ? 'Nothing cut here.' : 'Something slipped.'}</h1>
	<p>{page.status === 404 ? "That page doesn't exist." : page.error?.message}</p>
	<p><a href="/make">Make a template</a> or <a href="/patterns">browse patterns</a>.</p>
</div>

<style>
	.err {
		padding-block: var(--s8);
	}
</style>
