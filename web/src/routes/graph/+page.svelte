<script lang="ts">
import { goto } from "$app/navigation";
import { browser } from "$app/environment";
import GenreGraph from "$lib/GenreGraph.svelte";
import { buildGenreGraph } from "$lib/genre-graph";
import type { PageData } from "./$types";

let { data }: { data: PageData } = $props();

const SPARSE_THRESHOLD = 3;
const STORAGE_KEY = "graph:showSparse";
let showSparse = $state(browser ? localStorage.getItem(STORAGE_KEY) === "true" : false);

$effect(() => {
	localStorage.setItem(STORAGE_KEY, String(showSparse));
});

const visibleVertices = $derived(
	showSparse
		? data.vertices
		: data.vertices.filter((v) => v.artist_count >= SPARSE_THRESHOLD),
);
const visibleIds = $derived(new Set(visibleVertices.map((v) => v.genre_id)));
const visibleEdges = $derived(
	data.edges.filter(
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
	<span class="stat mono muted"
		>{data.vertices.length} genres · {data.edges.length} connections</span
	>
</div>

<GenreGraph {graph} onNodeTap={(genreId) => goto(`/genre/${genreId}`)} />

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

	.stat {
		margin-left: auto;
		font-size: 12px;
	}
</style>
