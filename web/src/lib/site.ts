// Site-wide constants for page metadata.

export const SITE = 'https://bushwhack.sstools.co';

/** Absolute URL for a path, with no trailing slash (except the root). */
export const canonical = (path: string) => SITE + (path === '/' ? '/' : path.replace(/\/+$/, ''));
