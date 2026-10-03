<script lang="ts">
	import '../app.css';
	import Header from '#lib/components/Header.svelte';
	import Footer from '#lib/components/Footer.svelte';

	let { children } = $props();
</script>

<a class="skip" href="#main">Skip to content</a>
<Header />
<main id="main">
	{@render children()}
</main>
<Footer />

<style>
	.skip {
		position: absolute;
		left: -999px;
		top: 0.5rem;
		background: var(--brass);
		color: var(--graphite);
		padding: 0.5rem 1rem;
		z-index: 10;
	}
	.skip:focus {
		left: 0.5rem;
	}
	main {
		flex: 1;
	}
</style>
