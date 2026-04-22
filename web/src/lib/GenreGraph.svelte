<script lang="ts">
import { onMount } from "svelte";
import { communityColor, edgeWidth, nodeSize } from "$lib/graph";
import type { GenreEdge, GenreVertex } from "../routes/api/graph/+server";
import type { Core, CytoscapeOptions } from "cytoscape";
import type { FcoseLayoutOptions } from "cytoscape-fcose";

let {
	vertices,
	edges,
	communities,
	onNodeTap,
}: {
	vertices: GenreVertex[];
	edges: GenreEdge[];
	communities: Map<number, number>;
	onNodeTap: (genreId: number) => void;
} = $props();

type CyLib = ((opts: CytoscapeOptions) => Core) & { use(ext: unknown): void };

let cyLib = $state<CyLib | null>(null);
let cyInstance: Core | null = null;
let graphEl: HTMLDivElement | undefined;
let loading = $state(true);

const numCommunities = $derived(Math.max(...communities.values(), 0) + 1);
const maxCount = $derived(Math.max(...vertices.map((v) => v.artist_count), 1));
const maxShared = $derived(
	Math.max(...edges.map((e) => e.shared_artist_count), 1),
);

const graphElements = $derived({
	nodes: vertices.map((v) => ({
		data: {
			id: String(v.genre_id),
			label: v.name,
			size: nodeSize(v.artist_count, maxCount),
			color: communityColor(communities.get(v.genre_id) ?? 0, numCommunities),
			genreId: v.genre_id,
		},
	})),
	edges: edges.map((e) => ({
		data: {
			source: String(e.source_genre_id),
			target: String(e.target_genre_id),
			width: edgeWidth(e.shared_artist_count, maxShared),
			opacity: 0.15 + 0.5 * Math.sqrt(e.shared_artist_count / maxShared),
			shared: e.shared_artist_count,
		},
	})),
});

onMount(() => {
	Promise.all([import("cytoscape"), import("cytoscape-fcose")]).then(
		([{ default: cytoscape }, { default: fcose }]) => {
			cytoscape.use(fcose);
			cyLib = cytoscape as unknown as CyLib;
		},
	);
	return () => cyInstance?.destroy();
});

$effect(() => {
	const elements = graphElements;
	const cy = cyLib;
	if (!cy || !graphEl) return;
	loading = true;
	const layout: FcoseLayoutOptions = {
		name: "fcose",
		animate: true,
		animationEasing: 'ease-out',
		quality: "proof",
		randomize: false,
		nodeRepulsion: () => 55000,
		idealEdgeLength: (edge) =>
			Math.max(50, 150 / Math.sqrt(edge.data("shared") || 1)),
		edgeElasticity: (edge) =>
			Math.min(0.9, 0.05 + (edge.data("shared") || 1) / 12),
		gravity: 0.65,
		gravityRange: 3.8,
		numIter: 2500,
		tile: true,
		tilingPaddingVertical: 10,
		tilingPaddingHorizontal: 10,
		fit: false,
		samplingType: true,
	};
	const id = setTimeout(() => {
		cyInstance?.destroy();
		cyInstance = cy({
			container: graphEl,
			elements,
			style: [
				{
					selector: "node",
					style: {
						"background-color": "data(color)",
						width: "data(size)",
						height: "data(size)",
						label: "data(label)",
						"font-size": "12px",
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
						"line-color": "rgb(96, 96, 160)",
						"line-opacity": "data(opacity)",
						"curve-style": "unbundled-bezier",
						"overlay-opacity": 0,
					},
				},
			] as CytoscapeOptions["style"],
			layout,
			minZoom: 0.5,
			maxZoom: 6,
			wheelSensitivity: 1.5,
			textureOnViewport: true,
			autoungrabify: false,
		});
		cyInstance?.on("tap", "node", (e) => {
			onNodeTap(e.target.data("genreId") as number);
		});
		loading = false;
	}, 16);
	return () => clearTimeout(id);
});
</script>

<div class="graph-wrapper card">
	{#if loading}
		<div class="overlay mono muted">Loading graph…</div>
	{/if}
	<div class="graph-container" bind:this={graphEl}></div>
</div>

<style>
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
