<script lang="ts">
import { onMount } from "svelte";
import { goto } from "$app/navigation";
import { detectCommunities } from "$lib/louvain";
import type { PageData } from "./$types";

let { data }: { data: PageData } = $props();

const SPARSE_THRESHOLD = 3;
let showSparse = $state(false);
let loading = $state(true);
let graphEl: HTMLDivElement;

// Module-level handles so toggle can reach them after mount
let cyLib: ((opts: unknown) => unknown) | null = null;
let cyInstance: {
	destroy(): void;
	on(
		evt: string,
		sel: string,
		fn: (e: { target: { data(k: string): unknown } }) => void,
	): void;
} | null = null;

// Precompute communities on the full dataset
const communities = $derived(
	detectCommunities(
		data.vertices.map((v) => v.genre_id),
		data.edges.map((e) => ({
			source: e.source_genre_id,
			target: e.target_genre_id,
			weight: e.shared_artist_count,
		})),
	),
);
const numCommunities = $derived(Math.max(...communities.values(), 0) + 1);

// Evenly-spaced HSL hues — offset by 200 to start near the accent purple
function communityColor(commId: number): string {
	const hue = Math.round(((commId / numCommunities) * 360 + 200) % 360);
	return `hsl(${hue}, 62%, 56%)`;
}

// Log-scale node diameter: 18–58 px
const maxCount = $derived(
	Math.max(...data.vertices.map((v) => v.artist_count), 1),
);
function nodeSize(count: number): number {
	return 18 + 40 * (Math.log(count + 1) / Math.log(maxCount + 1));
}

// Linear edge width: 0.5–4 px
const maxShared = $derived(
	Math.max(...data.edges.map((e) => e.shared_artist_count), 1),
);
function edgeWidth(count: number): number {
	return 0.5 + 3.5 * (count / maxShared);
}

function renderGraph(sparse: boolean) {
	if (!cyLib || !graphEl) return;
	cyInstance?.destroy();

	const verts = sparse
		? data.vertices
		: data.vertices.filter((v) => v.artist_count >= SPARSE_THRESHOLD);
	const visibleIds = new Set(verts.map((v) => v.genre_id));
	const filteredEdges = data.edges.filter(
		(e) =>
			visibleIds.has(e.source_genre_id) && visibleIds.has(e.target_genre_id),
	);

	// @ts-expect-error — dynamic import, no static type here
	cyInstance = cyLib({
		container: graphEl,
		elements: {
			nodes: verts.map((v) => ({
				data: {
					id: String(v.genre_id),
					label: v.name,
					size: nodeSize(v.artist_count),
					color: communityColor(communities.get(v.genre_id) ?? 0),
					genreId: v.genre_id,
				},
			})),
			edges: filteredEdges.map((e) => ({
				data: {
					source: String(e.source_genre_id),
					target: String(e.target_genre_id),
					width: edgeWidth(e.shared_artist_count),
				},
			})),
		},
		style: [
			{
				selector: "node",
				style: {
					"background-color": "data(color)",
					width: "data(size)",
					height: "data(size)",
					label: "data(label)",
					"font-size": "8px",
					"font-family": '"IBM Plex Sans", sans-serif',
					color: "#e2e2f0",
					"text-valign": "center",
					"text-halign": "center",
					"text-wrap": "wrap",
					"text-max-width": "data(size)",
					"min-zoomed-font-size": 7,
					"border-width": 1,
					"border-color": "rgba(0,0,0,0.35)",
					"overlay-opacity": 0,
					cursor: "pointer",
				},
			},
			{
				selector: "node:active",
				style: { "overlay-opacity": 0.12, "overlay-color": "#fff" },
			},
			{
				selector: "node:selected",
				style: { "border-width": 2, "border-color": "#7c6af7" },
			},
			{
				selector: "edge",
				style: {
					width: "data(width)",
					"line-color": "rgba(96, 96, 160, 0.35)",
					"curve-style": "haystack",
					"overlay-opacity": 0,
				},
			},
		],
		layout: {
			name: "cose",
			animate: false,
			randomize: false,
			nodeRepulsion: () => 6000,
			nodeOverlap: 20,
			idealEdgeLength: () => 80,
			edgeElasticity: () => 150,
			gravity: 0.4,
			numIter: 500,
			initialTemp: 180,
			coolingFactor: 0.99,
			minTemp: 1,
		},
		minZoom: 0.15,
		maxZoom: 6,
		wheelSensitivity: 0.3,
	});

	cyInstance?.on("click", "node", (e) => {
		goto(`/genre/${e.target.data("genreId")}`);
	});

	loading = false;
}

onMount(() => {
	import("cytoscape").then(({ default: cytoscape }) => {
		cyLib = cytoscape as unknown as typeof cyLib;
		renderGraph(showSparse);
	});
	return () => cyInstance?.destroy();
});

function toggleSparse() {
	showSparse = !showSparse;
	loading = true;
	// Small delay so the spinner renders before the synchronous layout runs
	setTimeout(() => renderGraph(showSparse), 16);
}
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
		<input type="checkbox" checked={showSparse} onchange={toggleSparse} />
		<span>Show genres with fewer than {SPARSE_THRESHOLD} artists</span>
	</label>
	<span class="stat mono muted">{data.vertices.length} genres · {data.edges.length} connections</span>
</div>

<div class="graph-wrapper card">
	{#if loading}
		<div class="overlay mono muted">Loading graph…</div>
	{/if}
	<div class="graph-container" bind:this={graphEl}></div>
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

	.graph-wrapper {
		position: relative;
		height: 75vh;
		min-height: 500px;
		padding: 0;
		overflow: hidden;
	}

	.overlay {
		position: absolute;
		inset: 0;
		display: flex;
		align-items: center;
		justify-content: center;
		font-size: 13px;
		z-index: 1;
		background: var(--surface);
	}

	.graph-container {
		width: 100%;
		height: 100%;
	}
</style>
