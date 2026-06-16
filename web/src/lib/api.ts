import type { OgData } from "./types";

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
