# Website maintenance

The site is a static Astro + Starlight build in `site/`. Read the official
[Astro deployment guide](https://docs.astro.build/en/guides/deploy/vercel/)
and [Starlight search guide](https://starlight.astro.build/guides/site-search/)
when changing these integrations.

## Local development

```sh
npm --prefix site ci
npm --prefix site run check
npm --prefix site run build
npm --prefix site run preview
```

Preview tests the built site, including Pagefind's build-time index. Restart a
development server after changing its environment variables.

## Vercel configuration

The `desk` project in `markos-narinians-projects` is configured for:

- Framework: Astro; root directory: `site`.
- Install: `npm ci`; build: `npm run build`; output: `dist`.
- Source files outside the root directory: enabled. The safety page imports the
  authoritative `docs/safety.md` from outside `site/`.
- GitHub source: `markosnarinian/desk`. Linking and Git-based deployments require
  the Vercel GitHub integration to have access to this private repository.

Creating the project alone does not deploy the site. Once linked, Git pushes can
trigger deployments. No server adapter is needed for this static site.

`SITE_URL`, when set, is the canonical production URL, including `https://`.
On Vercel, the config otherwise uses `VERCEL_PROJECT_PRODUCTION_URL` to generate
production canonical links and the sitemap, including for preview builds. Local
builds without either variable skip the sitemap.

## Algolia DocSearch

The official `@astrojs/starlight-docsearch` plugin is installed. It activates
when all three build environment variables are supplied:

| Variable | Value |
| --- | --- |
| `ALGOLIA_APP_ID` | Algolia application's ID |
| `ALGOLIA_SEARCH_API_KEY` | Public, search-only key scoped to the docs index |
| `ALGOLIA_INDEX_NAME` | A populated, DocSearch-compatible index |

These values reach the browser. **Never supply an Admin, crawler, write-capable,
or unrestricted key.** Restrict the search key to this index. Use Vercel's
environment settings or private credential entry, never a committed `.env` file.
Local builds read these values from the process environment, not `.env` files.

Before enabling:

1. Create an Algolia DocSearch index and configure its crawler for the production
   docs domain. The official [DocSearch program](https://docsearch.algolia.com/)
   is one way to obtain this setup; eligibility is determined by Algolia.
2. Crawl the docs, excluding the 404 page, and confirm records and heading links
   are populated. A key alone does not index pages.
3. Set the three variables for the intended Vercel environments and rebuild.
4. Search for `quarantine`, `SSH`, and `DESK_HOME`, then follow a result and check
   its target page and heading. Repeat after significant documentation changes.

With no Algolia variables, the site intentionally keeps working Pagefind search.
Partial Algolia configuration fails the build rather than silently activating a
broken modal. The integration is prepared, **not a claim that a hosted index has
been provisioned**. Do not disable the fallback until real search is configured.
