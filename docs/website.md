# Website maintenance

The site is a statically exported Next.js App Router app in `site/`, styled with
stock Chakra UI v3 and laid out after the Chakra UI Pro Docs Kit template.

## Layout

- `content/docs/*.mdx` and `content/commands/*.mdx`: page content. Front matter
  holds `title` and `description`.
- `lib/nav.ts`: unified sidebar groups and page ordering. Commands sit under
  Reference; their existing `/commands/*` URLs are preserved. A page appears
  on the site only when it is listed here.
- `/` renders the introduction directly, with `/docs/introduction/` as its
  canonical URL. Both routes use the same content and documentation layout.
- `components/mdx.tsx`: MDX components available in content: `Callout`,
  `Steps`/`Step`, `ParamField`, and `Tabs`/`Tab`. A fenced code
  block can take a label, such as ```` ```text title="Usage" ````.
- `app/search-index.json`: the client-side search index (⌘K), generated at build
  time from the same content.

## Local development

```sh
npm --prefix site ci
npm --prefix site run check
npm --prefix site run build
npm --prefix site run preview   # serves site/out on port 4321
npm --prefix site run dev
```

## Vercel configuration

The `desk` project in `markos-narinians-projects` uses root directory `site`.
`site/vercel.json` sets the framework to Next.js and the output directory to
`.next`, overriding the project's dashboard settings (still Astro's `dist`).
Vercel's Next.js builder reads `.next` and deploys the static export from it;
pointing it at `out` fails.
- GitHub source: `markosnarinian/desk`. Linking and Git-based deployments require
  the Vercel GitHub integration to have access to this private repository.

`SITE_URL`, when set, is the canonical production URL, including `https://`.
On Vercel, the site otherwise uses `VERCEL_PROJECT_PRODUCTION_URL` for canonical
links and the sitemap. Without either variable the sitemap is empty.
