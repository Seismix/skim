import type { OgData } from "./types";

export function getDomain(url: string): string {
  try {
    return new URL(url).hostname;
  } catch {
    return url;
  }
}

// Route external images through the local backend so they load same-origin.
// Browsers' tracking protection / hotlink checks otherwise block many og:images.
export function proxied(src: string): string {
  return /^https?:\/\//i.test(src) ? `/api/img?url=${encodeURIComponent(src)}` : src;
}

// An icon-sized og:image (e.g. a site's favicon used as og:image) is shown by
// Discord/Slack as the small provider icon, not a large embed image.
export function isIconImage(og: OgData): boolean {
  return (
    !!og.image &&
    (/favicon|apple-touch|\/icon|icon\.(?:png|ico|svg)/i.test(og.image) ||
      (!!og.imageWidth && parseInt(og.imageWidth, 10) < 200))
  );
}
