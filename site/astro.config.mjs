import { defineConfig } from 'astro/config';
import starlight from '@astrojs/starlight';
import starlightDocSearch from '@astrojs/starlight-docsearch';

const algolia = {
  appId: process.env.ALGOLIA_APP_ID,
  apiKey: process.env.ALGOLIA_SEARCH_API_KEY,
  indexName: process.env.ALGOLIA_INDEX_NAME,
};
const searchSettings = Object.values(algolia);
if (searchSettings.some(Boolean) && !searchSettings.every(Boolean)) {
  throw new Error('Algolia requires ALGOLIA_APP_ID, ALGOLIA_SEARCH_API_KEY, and ALGOLIA_INDEX_NAME together.');
}

export default defineConfig({
  site: process.env.SITE_URL || (process.env.VERCEL_PROJECT_PRODUCTION_URL
    ? `https://${process.env.VERCEL_PROJECT_PRODUCTION_URL}` : undefined),
  integrations: [starlight({
    title: 'desk',
    description: 'Managed local workspaces for the work in between.',
    favicon: '/favicon.svg',
    plugins: searchSettings.every(Boolean) ? [starlightDocSearch(algolia)] : [],
    customCss: ['./src/styles/custom.css'],
    components: {
      Hero: './src/components/Hero.astro',
      Header: './src/components/Header.astro',
    },
    sidebar: [
      { label: 'Start here', items: [
        { label: 'Installation', slug: 'installation' },
        { label: 'Quickstart', slug: 'quickstart' },
        { label: 'Cleanup guarantees', slug: 'safety' },
      ] },
      { label: 'Guides', items: [
        { label: 'Scratch workspaces', slug: 'guides/scratch' },
        { label: 'Repository workspaces', slug: 'guides/repositories' },
        { label: 'Cleanup and recovery', slug: 'guides/cleanup' },
      ] },
      { label: 'Reference', items: [
        { label: 'Command reference', slug: 'commands' },
        { label: 'Storage and configuration', slug: 'configuration' },
        { label: 'Troubleshooting', slug: 'troubleshooting' },
      ] },
    ],
  })],
});
