// @ts-check
import { defineConfig } from 'astro/config';
import starlight from '@astrojs/starlight';

// https://astro.build/config
export default defineConfig({
	site: 'https://go-gen-jsonschema.tylergannon.com',
	vite: {
		server: { fs: { allow: ['..', '../..'] } },
	},
	integrations: [
		starlight({
			title: 'go-gen-jsonschema',
			social: [{ icon: 'github', label: 'GitHub', href: 'https://github.com/tylergannon/go-gen-jsonschema' }],
			sidebar: [
				{
					label: 'Overview',
					items: [{ label: 'Introduction', slug: '' }],
				},
				{
					label: 'Spec',
					autogenerate: { directory: 'spec' },
				},
				{
					label: 'Examples',
					items: [{ label: 'Examples Index', slug: 'examples' }],
				},
				{
					label: 'Implementation',
					autogenerate: { directory: 'implementation' },
				},
				{
					label: 'API Reference',
					autogenerate: { directory: 'api' },
				},
			],
		}),
	],
});
