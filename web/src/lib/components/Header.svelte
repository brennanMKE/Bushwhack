<script lang="ts">
	import { page } from '$app/state';
	import Logo from './Logo.svelte';

	const links = [
		{ href: '/make', label: 'Make' },
		{ href: '/patterns', label: 'Patterns' },
		{ href: '/guide', label: 'Guide' }
	];
	const current = (href: string) =>
		page.url.pathname === href || page.url.pathname.startsWith(href + '/');
</script>

<header class="site-header">
	<div class="wrap bar">
		<a class="wordmark" href="/" aria-label="Bushwhack home"><Logo size={30} />Bushwhack</a>
		<nav aria-label="Main">
			{#each links as l (l.href)}
				<a href={l.href} aria-current={current(l.href) ? 'page' : undefined}>{l.label}</a>
			{/each}
		</nav>
	</div>
</header>

<style>
	.site-header {
		border-bottom: 1px solid var(--rule);
	}
	.bar {
		display: flex;
		align-items: center;
		justify-content: space-between;
		gap: var(--s2);
		min-height: 4rem;
	}
	.wordmark {
		display: inline-flex;
		align-items: center;
		gap: 0.55rem;
		font-family: var(--font-display);
		font-weight: 800;
		font-size: 1.9rem;
		letter-spacing: 0.01em;
		text-decoration: none;
		line-height: 1;
	}
	nav {
		display: flex;
		gap: clamp(0.9rem, 3vw, 2rem);
	}
	nav a {
		text-decoration: none;
		font-weight: 600;
		padding-block: 0.4rem;
		border-bottom: 3px solid transparent;
	}
	nav a:hover {
		border-bottom-color: color-mix(in srgb, var(--brass) 50%, transparent);
	}
	nav a[aria-current='page'] {
		border-bottom-color: var(--brass);
	}
	@media (max-width: 440px) {
		.wordmark {
			font-size: 1.5rem;
			gap: 0.4rem;
		}
		.wordmark :global(.logo) {
			width: 24px;
			height: 24px;
		}
		nav {
			gap: 0.85rem;
		}
		nav a {
			font-size: 0.95rem;
		}
	}
</style>
