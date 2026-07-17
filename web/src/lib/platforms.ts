// The platforms skim renders, in card order. `id` is the wire contract: it's what
// the API's `platforms` param accepts, so these ids are mirrored by platformIDs
// in main.go — TestPlatformRegistryInSync (there) fails the build if the lists
// drift, or if Results.svelte lacks a card snippet for an id listed here.
//
// `label` is the platform's short name; `card` is the fuller heading its preview
// card carries. Both are UI-only.
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
