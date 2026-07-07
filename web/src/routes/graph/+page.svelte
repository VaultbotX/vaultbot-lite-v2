<script lang="ts">
import { goto } from "$app/navigation";
import { browser } from "$app/environment";
import GenreGraph from "$lib/GenreGraph.svelte";
import { buildMixedGraph } from "$lib/mixed-graph";
import type { GraphData } from "../api/graph/+server";
import type { PageData } from "./$types";

let { data }: { data: PageData } = $props();

const ARTISTS_KEY = "graph:showArtists";
const DYNAMIC_KEY = "graph:showDynamic";

let showArtists = $state(browser ? localStorage.getItem(ARTISTS_KEY) !== "false" : true);
let showDynamic = $state(browser ? localStorage.getItem(DYNAMIC_KEY) === "true" : false);
let dynamicData = $state<GraphData | null>(null);
let loadingDynamic = $state(false);

$effect(() => {
	localStorage.setItem(ARTISTS_KEY, String(showArtists));
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

const activeData = $derived(showDynamic && dynamicData ? dynamicData : data);

const visibleArtistVertices = $derived(showArtists ? activeData.artistVertices : []);
const visibleGenreArtistEdges = $derived(showArtists ? activeData.genreArtistEdges : []);
const visibleArtistArtistEdges = $derived(showArtists ? activeData.artistArtistEdges : []);

const graph = $derived(
	buildMixedGraph(
		activeData.genreVertices,
		visibleArtistVertices,
		activeData.genreGenreEdges,
		visibleGenreArtistEdges,
		visibleArtistArtistEdges,
	),
);

const connectionCount = $derived(
	activeData.genreGenreEdges.length +
		visibleGenreArtistEdges.length +
		visibleArtistArtistEdges.length,
);

function handleNodeTap(id: number, kind: "genre" | "artist"): void {
	goto(kind === "genre" ? `/genres/${id}` : `/artists/${id}`);
}
</script>

<svelte:head>
	<title>Vaultbot :: Genre Graph</title>
</svelte:head>

<div class="page-header">
	<h1>Genre Graph</h1>
	<p class="muted">Explore connections between genres and artists in the archive. Click a node to drill down.</p>
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
			checked={showDynamic}
			onchange={() => (showDynamic = !showDynamic)}
		/>
		<span>Current playlist only <span class="pill">≤ 2 weeks</span></span>
	</label>
	<span class="stat mono muted"
		>{activeData.genreVertices.length} genres · {visibleArtistVertices.length} artists · {connectionCount} connections</span
	>
</div>

{#if loadingDynamic}
	<div class="loading card">Loading current playlist graph…</div>
{:else}
	<GenreGraph {graph} onNodeTap={handleNodeTap} />
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
		font-family: "IBM Plex Sans", monospace;
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
