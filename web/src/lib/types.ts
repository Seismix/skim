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
}

// One entry of a batch skim: the input URL paired with its data or a per-URL
// error. Mirrors the backend `fetchResult` shape.
export interface FetchResult {
    url: string;
    data?: OgData;
    error?: string;
}
