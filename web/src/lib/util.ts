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

// An icon-sized og:image (e.g. a site's favicon used as og:image) isn't shown as
// a large embed image: Discord/Slack demote it to the small provider icon,
// WhatsApp to a square thumb beside the text.
export function isIconImage(og: OgData): boolean {
    return (
        !!og.image &&
        (/favicon|apple-touch|\/icon|icon\.(?:png|ico|svg)/i.test(og.image) ||
            (!!og.imageWidth && parseInt(og.imageWidth, 10) < 200))
    );
}
