<script lang="ts">
import { untrack } from "svelte";
import ArtistDetailPanel from "#lib/ArtistDetailPanel.svelte";
import GenreDetailPanel from "#lib/GenreDetailPanel.svelte";
import type { SelectedNode, TimeRange } from "#lib/graph.ts";
import { detailFetchUrl } from "#lib/graph.ts";
import type { ArtistDetail } from "../routes/api/artists/[id]/+server";
import type { GenreDetail } from "../routes/api/genres/[id]/+server";

type Detail =
	| { kind: "genre"; data: GenreDetail }
	| { kind: "artist"; data: ArtistDetail };

let {
	selected,
	initialDetail,
	activeWindow,
	onSelect,
	onClose,
}: {
	selected: SelectedNode | null;
	initialDetail: Detail | null;
	activeWindow: TimeRange | null;
	onSelect: (id: number, kind: "genre" | "artist") => void;
	onClose: () => void;
} = $props();

// Seeded once from the initial prop value (the server-prefetched deep link);
// deliberately not reactive to later `initialDetail` changes — the effect
// below owns all subsequent updates. `untrack` marks that intent explicitly.
let detail = $state<Detail | null>(untrack(() => initialDetail));
let loading = $state(false);
let error = $state(false);

// `initialDetail` is only trustworthy for the very first non-null `selected`
// value (the parent seeds both together from the same deep-linked node), and
// only when there's no active time window — it was always fetched server-side
// as all-time data — it must never be reused once the user has clicked
// through to something else, or once a window is active.
let consumedInitial = false;

$effect(() => {
	const node = selected;
	const window = activeWindow;
	if (!node) {
		detail = null;
		return;
	}

	if (
		!consumedInitial &&
		!window &&
		initialDetail &&
		initialDetail.kind === node.kind
	) {
		consumedInitial = true;
		detail = initialDetail;
		return;
	}
	consumedInitial = true;

	loading = true;
	error = false;

	// Debounced so dragging the galaxy page's time-window slider (which fires
	// `oninput` continuously) doesn't flood the detail endpoint with a request
	// per tick — only the settled value triggers a fetch.
	const timeoutId = setTimeout(() => {
		fetch(detailFetchUrl(node, window))
			.then((r) => {
				if (!r.ok) throw new Error(`Failed to load: ${r.status}`);
				return r.json();
			})
			.then((data) => {
				detail =
					node.kind === "genre"
						? { kind: "genre", data }
						: { kind: "artist", data };
				loading = false;
			})
			.catch(() => {
				error = true;
				loading = false;
			});
	}, 250);

	return () => clearTimeout(timeoutId);
});
</script>

<aside class="drawer card" class:open={selected !== null}>
	{#if selected}
		<button type="button" class="close mono" onclick={onClose} aria-label="Close">✕</button>
		{#if loading}
			<div class="status muted mono">Loading…</div>
		{:else if error}
			<div class="status muted mono">Failed to load details.</div>
		{:else if detail?.kind === "genre"}
			<GenreDetailPanel
				data={detail.data}
				{activeWindow}
				onSelectGenre={(id) => onSelect(id, "genre")}
				onSelectArtist={(id) => onSelect(id, "artist")}
			/>
		{:else if detail?.kind === "artist"}
			<ArtistDetailPanel
				data={detail.data}
				{activeWindow}
				onSelectGenre={(id) => onSelect(id, "genre")}
				onSelectArtist={(id) => onSelect(id, "artist")}
			/>
		{/if}
	{/if}
</aside>

<style>
	.drawer {
		position: absolute;
		top: 0;
		right: 0;
		bottom: 0;
		width: 380px;
		max-width: 100%;
		overflow-y: auto;
		overflow-x: hidden;
		z-index: 5;
		pointer-events: none;
		opacity: 0;
		transform: translateX(100%);
		transition: transform 0.2s cubic-bezier(0.16, 1, 0.3, 1), opacity 0.15s ease;
		box-shadow: -4px 0 24px rgba(0, 0, 0, 0.3);
	}

	.drawer.open {
		pointer-events: auto;
		opacity: 1;
		transform: translateX(0);
	}

	.close {
		position: sticky;
		top: 0;
		float: right;
		background: none;
		border: none;
		cursor: pointer;
		color: var(--text-muted);
		font-size: 14px;
		padding: 0.25rem;
		margin: -0.25rem -0.25rem 0.5rem auto;
	}

	.close:hover {
		color: var(--accent);
	}

	.status {
		font-size: 13px;
		padding: 2rem 0;
		text-align: center;
	}

	@media (max-width: 900px) {
		.drawer.open {
			position: fixed;
			inset: 0;
			width: auto;
			z-index: 10;
			border-radius: 0;
			border-width: 0;
		}
	}
</style>
