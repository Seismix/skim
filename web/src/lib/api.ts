import type { FetchResult, OgData } from "./types";

export async function fetchOg(url: string): Promise<OgData> {
    const res = await fetch("/api/fetch-og", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ url }),
    });
    const data = await res.json();
    if (!res.ok) throw new Error(data.error || "Unknown error");
    return data as OgData;
}

// fetchOgBatch skims several URLs in one request. The server fetches them
// concurrently and returns one result per input URL — each with its own data
// or error — so a single bad link doesn't sink the whole batch.
export async function fetchOgBatch(urls: string[]): Promise<FetchResult[]> {
    const res = await fetch("/api/fetch-og", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ urls }),
    });
    const data = await res.json();
    if (!res.ok) throw new Error(data.error || "Unknown error");
    return (data.results || []) as FetchResult[];
}
