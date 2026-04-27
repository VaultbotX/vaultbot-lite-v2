<script lang="ts">
import { goto } from "$app/navigation";
import { browser } from "$app/environment";
import GenreGraph from "$lib/GenreGraph.svelte";
import { buildGenreGraph } from "$lib/genre-graph";
import type { GraphData } from "../api/graph/+server";
import type { PageData } from "./$types";

let { data }: { data: PageData } = $props();

const SPARSE_THRESHOLD = 3;
const SPARSE_KEY = "graph:showSparse";
const DYNAMIC_KEY = "graph:showDynamic";

let showSparse = $state(browser ? localStorage.getItem(SPARSE_KEY) === "true" : false);
let showDynamic = $state(browser ? localStorage.getItem(DYNAMIC_KEY) === "true" : false);
let dynamicData = $state<GraphData | null>(null);
let loadingDynamic = $state(false);

$effect(() => {
	localStorage.setItem(SPARSE_KEY, String(showSparse));
});

$effect(() => {
	localStorage.setItem(DYNAMIC_KEY, String(showDynamic));
});

$effect(() => {
	if (!showDynamic || dynamicData) return;
	loadingDynamic = true;
	fetch("/api/graph?dynamic=true")
		.then((r) => r.json() as Promise<GraphData>)
		.then((d) => {
			dynamicData = d;
			loadingDynamic = false;
		})
		.catch(() => {
			loadingDynamic = false;
		});
});

const activeVertices = $derived(showDynamic && dynamicData ? dynamicData.vertices : data.vertices);
const activeEdges = $derived(showDynamic && dynamicData ? dynamicData.edges : data.edges);

const visibleVertices = $derived(
	showSparse
		? activeVertices
		: activeVertices.filter((v) => v.artist_count >= SPARSE_THRESHOLD),
);
const visibleIds = $derived(new Set(visibleVertices.map((v) => v.genre_id)));
const visibleEdges = $derived(
	activeEdges.filter(
		(e) =>
			visibleIds.has(e.source_genre_id) && visibleIds.has(e.target_genre_id),
	),
);

const graph = $derived(buildGenreGraph(visibleVertices, visibleEdges));
</script>

<svelte:head>
	<title>Vaultbot — Genre Graph</title>
</svelte:head>

<div class="page-header">
	<h1>Genre Graph</h1>
	<p class="muted">Explore connections between genres in the archive. Click a node to drill down.</p>
</div>

<div class="toolbar card">
	<label>
		<input
			type="checkbox"
			checked={showSparse}
			onchange={() => (showSparse = !showSparse)}
		/>
		<span>Show genres with fewer than {SPARSE_THRESHOLD} artists</span>
	</label>
	<label>
		<input
			type="checkbox"
			checked={showDynamic}
			onchange={() => (showDynamic = !showDynamic)}
		/>
		<span>Current playlist only <span class="pill">≤ 2 weeks</span></span>
	</label>
	<span class="stat mono muted"
		>{visibleVertices.length} genres · {visibleEdges.length} connections</span
	>
</div>

{#if loadingDynamic}
	<div class="loading card">Loading current playlist graph…</div>
{:else}
	<GenreGraph {graph} onNodeTap={(genreId) => goto(`/genres/${genreId}`)} />
{/if}

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

	.pill {
		display: inline-block;
		font-family: "IBM Plex Mono", monospace;
		font-size: 10px;
		padding: 1px 5px;
		border-radius: 4px;
		background: var(--surface-2);
		border: 1px solid var(--border);
		color: var(--text-muted);
		vertical-align: middle;
	}

	.stat {
		margin-left: auto;
		font-size: 12px;
	}

	.loading {
		display: flex;
		align-items: center;
		justify-content: center;
		height: 400px;
		color: var(--text-muted);
		font-size: 14px;
	}
</style>
