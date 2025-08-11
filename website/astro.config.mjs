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
		}),
	],
});
