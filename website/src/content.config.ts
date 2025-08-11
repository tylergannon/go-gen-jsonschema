import { defineCollection } from 'astro:content';
import { docsLoader } from '@astrojs/starlight/loaders';
import { docsSchema } from '@astrojs/starlight/schema';

export const collections = {
	// Load docs from symlinked ../../docs via src/content/docs symlink
	docs: defineCollection({ loader: docsLoader(), schema: docsSchema() }),
};
