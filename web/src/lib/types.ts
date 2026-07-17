// What one platform renders for a page — the API's `platforms` block (mirrors
// platformRender in main.go). null means the platform doesn't render that field
// at all; "" means it would, but the page left it blank.
export interface PlatformRender {
    shape: string;
    title: string;
    description: string | null;
    image: string | null;
    domain: string;
    siteName: string | null;
}

export interface OgData {
    title: string;
    description: string;
    image: string;
    imageWidth: string;
    imageHeight: string;
    imageAlt: string;
    url: string;
    siteName: string;
    type: string;
    twitterCard: string;
    twitterSite: string;
    twitterCreator: string;
    locale: string;
    themeColor: string;
    favicon: string;
    allMeta: Record<string, string>;
    // Derived by the server, not extracted from the page: the og:image
    // classification the cards branch their layouts on, and each platform's
    // render. Both exist so the per-platform rules live only in main.go.
    imageKind: "none" | "icon" | "banner";
    platforms: Record<string, PlatformRender>;
}

// One entry of a batch skim: the input URL paired with its data or a per-URL
// error. Mirrors the backend `fetchResult` shape.
export interface FetchResult {
    url: string;
    data?: OgData;
    error?: string;
}
