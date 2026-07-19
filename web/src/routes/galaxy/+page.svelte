<script lang="ts">
import { untrack } from "svelte";
import { browser } from "$app/environment";
import { pushState } from "$app/navigation";
import { page } from "$app/state";
import GenreGraph from "$lib/GenreGraph.svelte";
import GraphDetailDrawer from "$lib/GraphDetailDrawer.svelte";
import type { SearchableNode, SelectedNode, TimeRange } from "$lib/graph";
import { formatWindowRange, rangesOverlap, searchNodes } from "$lib/graph";
import { buildMixedGraph } from "$lib/mixed-graph";
import type { GraphData } from "../api/graph/+server";
import type { PageData } from "./$types";

let { data }: { data: PageData } = $props();

const ARTISTS_KEY = "graph:showArtists";
const ALL_TIME_KEY = "graph:allTime";
const FOURTEEN_DAYS_SECONDS = 14 * 24 * 60 * 60;

function collectRanges(gd: GraphData): TimeRange[] {
	return [
		...gd.genreVertices.flatMap((v) => v.ranges),
		...gd.artistVertices.flatMap((v) => v.ranges),
	];
}

let showArtists = $state(
	browser ? localStorage.getItem(ARTISTS_KEY) !== "false" : true,
);
let allTime = $state(
	browser ? localStorage.getItem(ALL_TIME_KEY) !== "false" : true,
);

// The slider always represents a fixed 14-day-wide window; only its start
// (position within the archive's history) moves. Defaults to the most
// recent 14 days the first time the user leaves all-time mode, mirroring
// the old dynamic-mode default. Seeded once from the initial `data` load —
// there's no refetch to re-seed from, the graph never changes underneath.
let windowStart = $state(
	untrack(() => {
		const ranges = collectRanges(data);
		const max =
			ranges.length > 0 ? Math.max(...ranges.map(([, end]) => end)) : 0;
		return Math.max(0, max - FOURTEEN_DAYS_SECONDS);
	}),
);

let genreGraphInst = $state<{ focusNode: (nodeId: string) => void } | null>(
	null,
);
let searchQuery = $state("");
let activeIndex = $state(-1);
let searchWrapperEl: HTMLDivElement | undefined;

// Local state, not derived from the URL: `pushState` (shallow routing) never
// updates `page.url` — only `page.state` — so re-deriving selection from
// `page.url` would silently never react to a pivot. A plain `goto` would
// update `page.url`, but would also re-run this page's `load` function on
// every click, needlessly refetching the entire graph payload. Selection is
// therefore its own source of truth, kept in sync with `page.state` only for
// browser back/forward through shallow-routed history entries (see effect
// below).
let selectedNode = $state<SelectedNode | null>(untrack(() => data.initialNode));

$effect(() => {
	if ("node" in page.state) selectedNode = page.state.node ?? null;
});

$effect(() => {
	localStorage.setItem(ARTISTS_KEY, String(showArtists));
});

$effect(() => {
	localStorage.setItem(ALL_TIME_KEY, String(allTime));
});

// The full graph structure never changes with the window — only reducer
// visibility does (see GenreGraph's `activeWindow` prop) — so this is the
// only place the window is turned into a concrete [start, end] pair.
const activeWindow = $derived<TimeRange | null>(
	allTime ? null : [windowStart, windowStart + FOURTEEN_DAYS_SECONDS],
);

function isInWindow(ranges: TimeRange[]): boolean {
	const w = activeWindow;
	return !w || rangesOverlap(ranges, w[0], w[1]);
}

const allRanges = $derived(collectRanges(data));
const minTimestamp = $derived(
	allRanges.length > 0 ? Math.min(...allRanges.map(([start]) => start)) : 0,
);
const maxTimestamp = $derived(
	allRanges.length > 0 ? Math.max(...allRanges.map(([, end]) => end)) : 0,
);
// The window's start can't push its end past the latest known activity.
const windowStartMax = $derived(
	Math.max(minTimestamp, maxTimestamp - FOURTEEN_DAYS_SECONDS),
);

// The graph structure always contains every artist node/edge, regardless
// of `showArtists` — hiding them is GenreGraph's reducer's job (a redraw),
// not something that should rebuild the graph or rerun FA2/Louvain. These
// `showArtists`-filtered lists exist only for the stat line and search
// below, which should reflect what's actually visible.
const visibleArtistVertices = $derived(showArtists ? data.artistVertices : []);
const visibleGenreArtistEdges = $derived(
	showArtists ? data.genreArtistEdges : [],
);
const visibleArtistArtistEdges = $derived(
	showArtists ? data.artistArtistEdges : [],
);

const graph = $derived(
	buildMixedGraph(
		data.genreVertices,
		data.artistVertices,
		data.genreGenreEdges,
		data.genreArtistEdges,
		data.artistArtistEdges,
	),
);

// Distinct from the vertex/edge lists above: those feed the (always-full)
// graph structure, these are only for the stat line and search, which
// should reflect what's actually visible under the current window.
const genresInWindow = $derived(
	data.genreVertices.filter((v) => isInWindow(v.ranges)),
);
const artistsInWindow = $derived(
	visibleArtistVertices.filter((v) => isInWindow(v.ranges)),
);
const genreGenreEdgesInWindow = $derived(
	data.genreGenreEdges.filter((e) => isInWindow(e.ranges)),
);
const genreArtistEdgesInWindow = $derived(
	visibleGenreArtistEdges.filter((e) => isInWindow(e.ranges)),
);
const artistArtistEdgesInWindow = $derived(
	visibleArtistArtistEdges.filter((e) => isInWindow(e.ranges)),
);

const connectionCount = $derived(
	genreGenreEdgesInWindow.length +
		genreArtistEdgesInWindow.length +
		artistArtistEdgesInWindow.length,
);

const selectedGraphNodeId = $derived(
	selectedNode
		? `${selectedNode.kind === "genre" ? "g" : "a"}:${selectedNode.id}`
		: null,
);

const searchableNodes = $derived<SearchableNode[]>([
	...genresInWindow.map((v) => ({
		id: v.genre_id,
		kind: "genre" as const,
		name: v.name,
	})),
	...artistsInWindow.map((v) => ({
		id: v.artist_id,
		kind: "artist" as const,
		name: v.name,
	})),
]);

const searchResults = $derived(searchNodes(searchQuery, searchableNodes));
const searchOpen = $derived(searchQuery.trim().length > 0);

function selectSearchResult(node: SearchableNode): void {
	selectNode(node.id, node.kind);
	searchQuery = "";
	activeIndex = -1;
}

function handleSearchKeydown(e: KeyboardEvent): void {
	if (searchResults.length === 0) {
		if (e.key === "Escape") {
			searchQuery = "";
			activeIndex = -1;
		}
		return;
	}
	if (e.key === "ArrowDown") {
		e.preventDefault();
		activeIndex = (activeIndex + 1) % searchResults.length;
	} else if (e.key === "ArrowUp") {
		e.preventDefault();
		activeIndex =
			(activeIndex - 1 + searchResults.length) % searchResults.length;
	} else if (e.key === "Enter") {
		e.preventDefault();
		selectSearchResult(searchResults[Math.max(activeIndex, 0)]);
	} else if (e.key === "Escape") {
		searchQuery = "";
		activeIndex = -1;
	}
}

function handleWindowClick(e: MouseEvent): void {
	if (searchWrapperEl && !searchWrapperEl.contains(e.target as Node)) {
		searchQuery = "";
		activeIndex = -1;
	}
}

function selectNode(id: number, kind: "genre" | "artist"): void {
	// pushState's second argument must be structured-cloneable by the History
	// API, so a plain object literal is passed here rather than `selectedNode`
	// itself — that's a $state reactive Proxy, which history.pushState throws
	// a DataCloneError on and silently aborts the URL update.
	selectedNode = { kind, id };
	const prefix = kind === "genre" ? "g" : "a";
	pushState(`?node=${prefix}:${id}`, { node: { kind, id } });
	// Single chokepoint for every selection path (canvas click, drawer
	// cross-link pivot, search result) so the camera always frames whatever
	// node is currently selected, not just search-originated ones.
	genreGraphInst?.focusNode(`${prefix}:${id}`);
}

function clearSelection(): void {
	selectedNode = null;
	pushState(page.url.pathname, { node: null });
}
</script>

<svelte:head>
	<title>Vaultbot :: Playlist Galaxy</title>
</svelte:head>

<svelte:window onclick={handleWindowClick} />

<div class="page-header">
	<h1>Playlist Galaxy</h1>
</div>

<div class="toolbar card">
	<label>
		<input
			type="checkbox"
			checked={showArtists}
			onchange={() => (showArtists = !showArtists)}
		/>
		<span>Show artists</span>
	</label>
	<label>
		<input
			type="checkbox"
			checked={allTime}
			onchange={() => (allTime = !allTime)}
		/>
		<span>All time</span>
	</label>
	{#if !allTime}
		<div class="window-control">
			<input
				type="range"
				class="window-slider"
				min={minTimestamp}
				max={windowStartMax}
				step={3600}
				value={windowStart}
				oninput={(e) => (windowStart = Number(e.currentTarget.value))}
			/>
			<span class="window-label mono muted"
				>{formatWindowRange(windowStart, windowStart + FOURTEEN_DAYS_SECONDS)}</span
			>
		</div>
	{/if}
	<div class="search-wrapper" bind:this={searchWrapperEl}>
		<input
			type="text"
			class="search-input"
			placeholder="Find a genre or artist…"
			bind:value={searchQuery}
			onkeydown={handleSearchKeydown}
		/>
		{#if searchOpen}
			<div class="search-dropdown card">
				{#if searchResults.length === 0}
					<div class="search-empty muted">No matches</div>
				{:else}
					{#each searchResults as result, i (`${result.kind}:${result.id}`)}
						<button
							type="button"
							class="search-result"
							class:active={i === activeIndex}
							onclick={() => selectSearchResult(result)}
							onmouseenter={() => (activeIndex = i)}
						>
							<span class="search-glyph"
								>{result.kind === "artist" ? "🎨" : "🎵"}</span
							>
							<span class="search-name">{result.name}</span>
						</button>
					{/each}
				{/if}
			</div>
		{/if}
	</div>
	<span class="stat mono muted"
		>{genresInWindow.length} genres · {artistsInWindow.length} artists · {connectionCount} connections</span
	>
</div>

<div class="graph-row">
	<GenreGraph
		bind:this={genreGraphInst}
		{graph}
		selectedNode={selectedGraphNodeId}
		{activeWindow}
		{showArtists}
		onNodeTap={selectNode}
		onBackgroundClick={clearSelection}
	/>
	<GraphDetailDrawer
		selected={selectedNode}
		initialDetail={data.initialDetail}
		{activeWindow}
		onSelect={selectNode}
		onClose={clearSelection}
	/>
</div>

<style>
	.page-header {
		margin-bottom: 1.5rem;
	}

	.page-header h1 {
		font-size: 24px;
		margin-bottom: 0.25rem;
	}

	.toolbar {
		margin-bottom: 1rem;
		display: flex;
		align-items: center;
		gap: 1rem;
		padding: 0.75rem 1rem;
		flex-wrap: wrap;
	}

	.toolbar label {
		display: flex;
		align-items: center;
		gap: 0.5rem;
		cursor: pointer;
		color: var(--text-muted);
		font-size: 13px;
		user-select: none;
	}

	.toolbar input[type="checkbox"] {
		accent-color: var(--accent);
		cursor: pointer;
	}

	.window-control {
		display: flex;
		align-items: center;
		gap: 0.5rem;
	}

	.window-slider {
		accent-color: var(--accent);
		cursor: pointer;
		width: 160px;
	}

	.window-label {
		font-size: 11px;
		white-space: nowrap;
	}

	.stat {
		margin-left: auto;
		font-size: 12px;
	}

	.search-wrapper {
		position: relative;
		width: 240px;
	}

	.search-input {
		width: 100%;
		background: var(--surface-2);
		border: 1px solid var(--border);
		border-radius: var(--radius);
		color: var(--text);
		font-size: 13px;
		padding: 6px 10px;
	}

	.search-input:focus {
		outline: none;
		border-color: var(--accent);
	}

	.search-dropdown {
		position: absolute;
		top: calc(100% + 4px);
		left: 0;
		right: 0;
		z-index: 10;
		padding: 4px;
		max-height: 280px;
		overflow-y: auto;
	}

	.search-empty {
		padding: 8px 10px;
		font-size: 13px;
	}

	.search-result {
		display: flex;
		align-items: center;
		gap: 8px;
		width: 100%;
		padding: 6px 10px;
		border-radius: calc(var(--radius) - 4px);
		background: transparent;
		color: var(--text);
		font-size: 13px;
		text-align: left;
		cursor: pointer;
	}

	.search-result:hover,
	.search-result.active {
		background: var(--surface-2);
	}

	.search-glyph {
		font-size: 12px;
	}

	.search-name {
		overflow: hidden;
		text-overflow: ellipsis;
		white-space: nowrap;
	}

	.graph-row {
		display: flex;
		align-items: stretch;
		gap: 1rem;
		height: 75vh;
		min-height: 500px;
	}
</style>
