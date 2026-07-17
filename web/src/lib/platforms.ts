// The platforms skim renders, in card order. `id` is the wire contract: it's what
// ?platforms= carries and what the API's `platforms` param accepts, so these ids
// are mirrored by platformIDs in main.go — TestPlatformRegistryInSync (there)
// fails the build if the lists drift, or if Results.svelte lacks a card snippet
// for an id listed here.
//
// `label` is the short name for the toggle row; `card` is the fuller heading its
// preview card carries. Both are UI-only.
export const PLATFORMS = [
    { id: "facebook", label: "Facebook", card: "Facebook / Open Graph" },
    { id: "twitter", label: "Twitter/X", card: "Twitter / X" },
    { id: "linkedin", label: "LinkedIn", card: "LinkedIn" },
    { id: "discord", label: "Discord", card: "Discord" },
    { id: "slack", label: "Slack", card: "Slack" },
    { id: "whatsapp", label: "WhatsApp", card: "WhatsApp" },
    { id: "imessage", label: "iMessage", card: "iMessage" },
] as const;

export type PlatformId = (typeof PLATFORMS)[number]["id"];

export const ALL_IDS: PlatformId[] = PLATFORMS.map((p) => p.id);

function isPlatformId(v: string): v is PlatformId {
    return (ALL_IDS as string[]).includes(v);
}

// Read the selection out of ?platforms=. Absent means every platform, so the
// default case keeps a clean URL; present-but-empty means none, which is how an
// all-off selection survives a reload. Unknown ids are dropped rather than
// rejected — the address bar is hand-editable, so a typo shouldn't be fatal here
// (the API is strict about the same input; see selectPlatforms in main.go).
export function parsePlatformParam(search: string): PlatformId[] {
    const raw = new URLSearchParams(search).get("platforms");
    if (raw === null) return [...ALL_IDS];
    return raw
        .split(",")
        .map((s) => s.trim().toLowerCase())
        .filter(isPlatformId);
}

// Serialize a selection for the address bar, always in PLATFORMS order so the URL
// is stable however the toggles were clicked. Returns null when everything is on,
// meaning the param should be dropped entirely rather than spelled out.
export function serializePlatformParam(selected: PlatformId[]): string | null {
    if (selected.length === ALL_IDS.length) return null;
    return ALL_IDS.filter((id) => selected.includes(id)).join(",");
}
