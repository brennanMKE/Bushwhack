<script lang="ts">
	let { size = 32 }: { size?: number } = $props();
</script>

<!-- Bushing in plan view: brass ring, bit dot, a thin chalk arc outside. -->
<svg class="logo" width={size} height={size} viewBox="0 0 32 32" aria-hidden="true">
	<path d="M5.2 9.8A13 13 0 0 1 22.2 4.6" fill="none" stroke="var(--chalk)" stroke-width="1.6" stroke-linecap="round" />
	<circle cx="16" cy="16" r="10" fill="none" stroke="var(--brass)" stroke-width="4" />
	<circle cx="16" cy="16" r="4" fill="currentColor" />
</svg>

<style>
	.logo {
		flex: none;
	}
</style>
