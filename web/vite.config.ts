import { defineConfig } from 'vitest/config';
import adapter from '@sveltejs/adapter-static';
import { sveltekit } from '@sveltejs/kit/vite';

export default defineConfig({
	plugins: [
		sveltekit({
			compilerOptions: {
				runes: ({ filename }) =>
					filename.split(/[/\\]/).includes('node_modules') ? undefined : true
			},
			// Content pages are prerendered to <route>.html; every other route
			// (the workbench, pattern pages) loads the SPA shell, spa.html. The
			// Go binary serves both.
			adapter: adapter({ pages: 'build', assets: 'build', fallback: 'spa.html', strict: false })
		})
	],
	server: {
		proxy: { '/api': 'http://127.0.0.1:8080', '/healthz': 'http://127.0.0.1:8080' }
	},
	test: {
		expect: { requireAssertions: true },
		include: ['src/**/*.{test,spec}.{js,ts}']
	}
});
