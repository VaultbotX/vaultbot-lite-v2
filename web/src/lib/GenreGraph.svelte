<script lang="ts">
import type Graph from "graphology";
import { onMount } from "svelte";

let {
	graph,
	onNodeTap,
}: {
	graph: Graph;
	onNodeTap: (genreId: number) => void;
} = $props();

// Structural types for dynamically-imported renderer libs — avoids bundling
// sigma/FA2 during SSR while still giving us precise type safety at call sites.
type SigmaInst = {
	kill(): void;
	on(event: string, cb: (payload: Record<string, unknown>) => void): void;
};
type SigmaLib = {
	new (
		graph: Graph,
		container: HTMLElement,
		settings?: Record<string, unknown>,
	): SigmaInst;
};
type FA2Lib = {
	assign(
		graph: Graph,
		opts: {
			iterations: number;
			settings?: Record<string, unknown>;
			getEdgeWeight?: string;
		},
	): void;
	inferSettings(graph: Graph): Record<string, unknown>;
};
type RandomLib = { assign(graph: Graph, opts?: { scale?: number }): void };

let sigmaLib = $state<SigmaLib | null>(null);
let fa2Lib = $state<FA2Lib | null>(null);
let randomLib = $state<RandomLib | null>(null);
let sigmaInst: SigmaInst | null = null;
let containerEl: HTMLDivElement | undefined;
let loading = $state(true);

onMount(() => {
	Promise.all([
		import("sigma"),
		import("graphology-layout-forceatlas2"),
		import("graphology-layout/random"),
	]).then(([s, f, r]) => {
		sigmaLib = s.default as unknown as SigmaLib;
		fa2Lib = f.default as unknown as FA2Lib;
		randomLib = r.default as unknown as RandomLib;
	});
	return () => sigmaInst?.kill();
});

$effect(() => {
	const g = graph;
	const Sigma = sigmaLib;
	const fa2 = fa2Lib;
	const random = randomLib;
	const container = containerEl;
	if (!Sigma || !fa2 || !random || !container) return;

	loading = true;

	const id = setTimeout(() => {
		// Assign random initial positions, then run ForceAtlas2
		random.assign(g, { scale: 200 });
		fa2.assign(g, {
			iterations: 500,
			settings: {
				...fa2.inferSettings(g),
				gravity: 1,
				scalingRatio: 10,
				adjustSizes: true,
			},
			getEdgeWeight: "shared",
		});

		sigmaInst?.kill();
		sigmaInst = new Sigma(g, container, {
			renderEdgeLabels: false,
			labelFont: "'IBM Plex Sans', sans-serif",
			labelSize: 12,
			labelWeight: "normal",
			labelColor: { color: "#e2e2f0" },
			labelRenderedSizeThreshold: 8,
			minCameraRatio: 0.05,
			maxCameraRatio: 8,
			stagePadding: 40,
			defaultEdgeColor: "rgb(96, 96, 160)",
			defaultNodeColor: "#7c6af7",
		});

		sigmaInst.on("clickNode", (payload) => {
			const node = payload.node as string;
			onNodeTap(g.getNodeAttribute(node, "genreId") as number);
		});

		sigmaInst.on("enterNode", () => {
			container.style.cursor = "pointer";
		});

		sigmaInst.on("leaveNode", () => {
			container.style.cursor = "default";
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
	<div class="graph-container" bind:this={containerEl}></div>
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
