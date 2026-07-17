export function getDomain(url: string): string {
    try {
        return new URL(url).hostname;
    } catch {
        return url;
    }
}

// Route external images through the local backend so they load same-origin.
// Browsers' tracking protection / hotlink checks otherwise block many og:images.
// (The icon-image heuristic used to live here too; it's now isIconImage in
// main.go only, and arrives pre-computed as OgData.imageKind.)
export function proxied(src: string): string {
    return /^https?:\/\//i.test(src) ? `/api/img?url=${encodeURIComponent(src)}` : src;
}
