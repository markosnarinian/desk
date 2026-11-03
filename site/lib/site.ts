/** Canonical production origin; on Vercel, preview builds also point at production. */
export const siteUrl =
  process.env.SITE_URL ||
  (process.env.VERCEL_PROJECT_PRODUCTION_URL ? `https://${process.env.VERCEL_PROJECT_PRODUCTION_URL}` : undefined)
