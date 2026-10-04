// Client-rendered by default; content pages opt into prerendering in their
// +page.ts. The Go server serves spa.html for routes that are not prerendered.
export const ssr = false;
export const prerender = false;
export const trailingSlash = 'never';
