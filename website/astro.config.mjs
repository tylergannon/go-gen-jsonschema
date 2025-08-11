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
				description: 'Generate JSON Schema from Go types with confidence. Enums, interfaces, providers, deterministic outputs.',
				logo: { src: './src/assets/gopher-front.svg', replacesTitle: true },
				favicon: '/favicon.svg',
				social: [
					{ icon: 'github', label: 'GitHub', href: 'https://github.com/tylergannon/go-gen-jsonschema' },
				],
				customCss: [
					'./src/fonts/font-face.css',
					'./src/styles/custom.css'
				],
				sidebar: [
					{ label: 'Getting Started', link: '/getting-started/' },
					{ label: 'Spec', link: '/spec/' },
					{ label: 'Examples', link: '/examples/' },
					{ label: 'API', link: '/api/' },
					{ label: 'Guides', autogenerate: { directory: 'guides' } },
					{
						label: 'Overview',
						items: [
							{ label: 'Home', link: '/' },
							{ label: 'Getting Started', link: '/getting-started/' },
						],
					},
					{
						label: 'Guides',
						autogenerate: { directory: 'guides' },
					},
					{ label: 'Spec', link: '/spec/' },
					{ label: 'Examples', link: '/examples/' },
					{
						label: 'API Reference',
						items: [ { label: 'Index', link: '/api/' } ],
					},
				],
			}),	],
});
