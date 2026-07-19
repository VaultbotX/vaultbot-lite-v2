<script lang="ts">
import type Graph from "graphology";
import { onMount } from "svelte";
import { isolatedNodePosition, rangesOverlap, type TimeRange } from "./graph";

let {
	graph,
	selectedNode,
	activeWindow,
	showArtists,
	onNodeTap,
	onBackgroundClick,
}: {
	graph: Graph;
	selectedNode: string | null;
	activeWindow: TimeRange | null;
	showArtists: boolean;
	onNodeTap: (id: number, kind: "genre" | "artist") => void;
	onBackgroundClick: () => void;
} = $props();

// Structural types for dynamically-imported renderer libs — avoids bundling
// sigma/FA2 during SSR while still giving us precise type safety at call sites.
type SigmaInst = {
	kill(): void;
	refresh(): void;
	on(event: string, cb: (payload: Record<string, unknown>) => void): void;
	getCamera(): {
		animate(
			state: Record<string, unknown>,
			opts: Record<string, unknown>,
		): void;
	};
	getNodeDisplayData(node: string): { x: number; y: number } | undefined;
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
type NodeBorderLib = {
	createNodeBorderProgram(options: {
		borders: Array<{
			size: { value: number; mode?: "relative" | "pixels" } | { fill: true };
			color:
				| { value: string }
				| { attribute: string; defaultValue?: string }
				| { transparent: true };
		}>;
	}): unknown;
};
type SigmaRenderingLib = {
	drawDiscNodeLabel(
		context: CanvasRenderingContext2D,
		data: Record<string, unknown>,
		settings: Record<string, unknown>,
	): void;
};

let sigmaLib = $state<SigmaLib | null>(null);
let fa2Lib = $state<FA2Lib | null>(null);
let edgeCurveLib = $state<unknown>(null);
let nodeBorderLib = $state<NodeBorderLib | null>(null);
let sigmaRenderingLib = $state<SigmaRenderingLib | null>(null);
let sigmaInst: SigmaInst | null = null;
let containerEl: HTMLDivElement | undefined;
let loading = $state(true);

// Hover and selection state — closed over by the Sigma reducers/event
// handlers below. Hover always takes visual priority over selection; when
// hover ends, the selection's highlight (if any) reappears.
let hoveredNode: string | null = null;
let hoveredNeighborSet = new Set<string>();
let selectedNeighborSet = new Set<string>();

onMount(() => {
	Promise.all([
		import("sigma"),
		import("graphology-layout-forceatlas2"),
		import("@sigma/edge-curve"),
		import("@sigma/node-border"),
		import("sigma/rendering"),
	]).then(([s, f, ec, nb, sr]) => {
		sigmaLib = s.default as unknown as SigmaLib;
		fa2Lib = f.default as unknown as FA2Lib;
		edgeCurveLib = ec.default;
		nodeBorderLib = nb as unknown as NodeBorderLib;
		sigmaRenderingLib = sr as unknown as SigmaRenderingLib;
	});
	return () => sigmaInst?.kill();
});

/**
 * Place each community's nodes in a small circle around a community center,
 * arranged on a larger ring. FA2 then refines from this warm start rather
 * than having to discover community structure from a random scatter.
 */
function initCommunityLayout(g: Graph): void {
	const groups = new Map<number, string[]>();
	g.forEachNode((node, attrs) => {
		const c = attrs.community as number;
		const group = groups.get(c) ?? [];
		groups.set(c, group);
		group.push(node);
	});

	const commList = [...groups.keys()];
	const centerR = 600;
	const memberR = 80;

	// Golden-angle spiral: fills a disk evenly with no ring/hollow-center artifact.
	const GOLDEN_ANGLE = Math.PI * (3 - Math.sqrt(5)); // ≈ 137.5°
	const communityCenters = new Map<number, { x: number; y: number }>();
	commList.forEach((commId, ci) => {
		const r = centerR * Math.sqrt(ci / commList.length);
		communityCenters.set(commId, {
			x: r * Math.cos(ci * GOLDEN_ANGLE),
			y: r * Math.sin(ci * GOLDEN_ANGLE),
		});
	});

	commList.forEach((commId) => {
		const members = groups.get(commId) ?? [];
		const center = communityCenters.get(commId) ?? { x: 0, y: 0 };
		members.forEach((node, mi) => {
			// Isolated (degree-0) nodes have no edges to pull them back toward
			// their community during FA2, so anchor them at the community
			// center and mark them fixed rather than letting them drift away.
			if (g.degree(node) === 0) {
				const pos = isolatedNodePosition(commId, communityCenters);
				g.setNodeAttribute(node, "x", pos.x);
				g.setNodeAttribute(node, "y", pos.y);
				g.setNodeAttribute(node, "fixed", true);
				return;
			}
			const a = (2 * Math.PI * mi) / members.length;
			g.setNodeAttribute(node, "x", center.x + memberR * Math.cos(a));
			g.setNodeAttribute(node, "y", center.y + memberR * Math.sin(a));
		});
	});
}

$effect(() => {
	const g = graph;
	const Sigma = sigmaLib;
	const fa2 = fa2Lib;
	const container = containerEl;
	const edgeCurve = edgeCurveLib;
	const nodeBorder = nodeBorderLib;
	const sigmaRendering = sigmaRenderingLib;
	if (
		!Sigma ||
		!fa2 ||
		!edgeCurve ||
		!container ||
		!nodeBorder ||
		!sigmaRendering
	)
		return;

	loading = true;

	const { drawDiscNodeLabel } = sigmaRendering;

	// Default (non-hover) label renderer: draws a genre/artist glyph centered
	// on the node itself, plus the name label as solid white text over a
	// near-opaque black stroke (no blur) — maximum contrast against the dense
	// web of edges, versus sigma's plain flat-color text.
	function drawNodeLabelWithOutline(
		context: CanvasRenderingContext2D,
		data: Record<string, unknown>,
		settings: Record<string, unknown>,
	): void {
		if (!data.label) return;
		const size = settings.labelSize as number;
		const font = settings.labelFont as string;
		const weight = settings.labelWeight as string;
		const x = data.x as number;
		const y = data.y as number;
		const nodeSize = data.size as number;
		const label = data.label as string;
		const kind = data.kind as string;

		context.save();

		const emoji = kind === "artist" ? "🎨" : "🎵";
		const emojiSize = Math.max(8, nodeSize * 1.2);
		context.font = `${emojiSize}px sans-serif`;
		context.textAlign = "center";
		context.textBaseline = "middle";
		context.fillText(emoji, x, y);

		context.textAlign = "left";
		context.textBaseline = "alphabetic";
		context.font = `${weight} ${size}px ${font}`;
		context.lineJoin = "round";
		context.lineWidth = 3;
		context.strokeStyle = "rgba(0, 0, 0, 0.95)";
		context.fillStyle = "#ffffff";
		const labelX = x + nodeSize + 3;
		const labelY = y + size / 3;
		context.strokeText(label, labelX, labelY);
		context.fillText(label, labelX, labelY);

		context.restore();
	}

	// Border program: 3px black outer ring, community color fill.
	// Border color is read from the `borderColor` attribute so the nodeReducer
	// can dim it alongside the fill when a neighbor is not highlighted.
	const nodeBorderProgram = nodeBorder.createNodeBorderProgram({
		borders: [
			{
				size: { value: 3, mode: "pixels" },
				color: { attribute: "borderColor", defaultValue: "#000000" },
			},
			{
				size: { fill: true },
				color: { attribute: "color", defaultValue: "#7c6af7" },
			},
		],
	});

	// Custom hover renderer: same pill shape as Sigma's default but with a dark
	// background so the white label text is readable against `--surface`.
	function drawDarkNodeHover(
		context: CanvasRenderingContext2D,
		data: Record<string, unknown>,
		settings: Record<string, unknown>,
	): void {
		const size = settings.labelSize as number;
		const font = settings.labelFont as string;
		const weight = settings.labelWeight as string;
		context.font = `${weight} ${size}px ${font}`;

		context.fillStyle = "#131318";
		context.shadowOffsetX = 0;
		context.shadowOffsetY = 0;
		context.shadowBlur = 8;
		context.shadowColor = "#000";

		const PADDING = 2;
		const label = data.label as string | null;
		const x = data.x as number;
		const y = data.y as number;
		const nodeSize = data.size as number;

		if (typeof label === "string") {
			const textWidth = context.measureText(label).width;
			const boxWidth = Math.round(textWidth + 5);
			const boxHeight = Math.round(size + 2 * PADDING);
			const radius = Math.max(nodeSize, size / 2) + PADDING;
			const angleRadian = Math.asin(boxHeight / 2 / radius);
			const xDeltaCoord = Math.sqrt(
				Math.abs(radius ** 2 - (boxHeight / 2) ** 2),
			);

			context.beginPath();
			context.moveTo(x + xDeltaCoord, y + boxHeight / 2);
			context.lineTo(x + radius + boxWidth, y + boxHeight / 2);
			context.lineTo(x + radius + boxWidth, y - boxHeight / 2);
			context.lineTo(x + xDeltaCoord, y - boxHeight / 2);
			context.arc(x, y, radius, angleRadian, -angleRadian);
			context.closePath();
			context.fill();
		} else {
			context.beginPath();
			context.arc(x, y, nodeSize + PADDING, 0, Math.PI * 2);
			context.closePath();
			context.fill();
		}

		context.shadowOffsetX = 0;
		context.shadowOffsetY = 0;
		context.shadowBlur = 0;

		// The hover renderer replaces the default label renderer entirely for the
		// hovered node, so the kind glyph has to be redrawn here too — otherwise
		// it visibly disappears the moment a node is hovered.
		const kind = data.kind as string;
		const emoji = kind === "artist" ? "🎨" : "🎵";
		const emojiSize = Math.max(8, nodeSize * 1.2);
		context.save();
		context.font = `${emojiSize}px sans-serif`;
		context.textAlign = "center";
		context.textBaseline = "middle";
		context.fillText(emoji, x, y);
		context.restore();

		drawDiscNodeLabel(context, data, settings);
	}

	const id = setTimeout(() => {
		initCommunityLayout(g);
		fa2.assign(g, {
			iterations: 500,
			settings: {
				...fa2.inferSettings(g),
				gravity: 1,
				scalingRatio: 10,
				adjustSizes: true,
				// Barnes-Hut approximation (O(n log n) per iteration) instead of the
				// exact O(n²) pairwise repulsion — with ~1,300 mixed genre/artist
				// nodes, the exact computation was blocking the main thread for
				// several seconds before the graph could even paint.
				barnesHutOptimize: true,
			},
			getEdgeWeight: "weight",
		});

		let hoverTimer: ReturnType<typeof setTimeout> | null = null;
		selectedNeighborSet =
			selectedNode && g.hasNode(selectedNode)
				? new Set(g.neighbors(selectedNode))
				: new Set();

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
			defaultEdgeType: "curve",
			defaultDrawNodeLabel: drawNodeLabelWithOutline,
			defaultDrawNodeHover: drawDarkNodeHover,
			edgeProgramClasses: { curve: edgeCurve },
			nodeProgramClasses: { circle: nodeBorderProgram },
			// Dim nodes/edges outside the active (hovered, or else selected) node's
			// neighborhood. Hover takes priority over selection while it's active.
			// A node/edge outside the active time window, or an artist node/edge
			// while artists are toggled off, is hidden outright, ahead of and
			// regardless of hover/selection dimming.
			nodeReducer: (node: unknown, data: unknown) => {
				const d = data as Record<string, unknown>;
				if (!showArtists && d.kind === "artist") {
					return { ...d, hidden: true };
				}
				if (activeWindow) {
					const ranges = (d.ranges as TimeRange[] | undefined) ?? [];
					if (!rangesOverlap(ranges, activeWindow[0], activeWindow[1])) {
						return { ...d, hidden: true };
					}
				}
				const activeNode = hoveredNode ?? selectedNode;
				const activeNeighbors = hoveredNode
					? hoveredNeighborSet
					: selectedNeighborSet;
				if (
					!activeNode ||
					node === activeNode ||
					activeNeighbors.has(node as string)
				) {
					return d;
				}
				return { ...d, color: "#1e1e28", borderColor: "#1e1e28", label: "" };
			},
			edgeReducer: (edge: unknown, data: unknown) => {
				const d = data as Record<string, unknown>;
				if (!showArtists && d.kind !== "genre-genre") {
					return { ...d, hidden: true };
				}
				if (activeWindow) {
					const ranges = (d.ranges as TimeRange[] | undefined) ?? [];
					if (!rangesOverlap(ranges, activeWindow[0], activeWindow[1])) {
						return { ...d, hidden: true };
					}
				}
				const activeNode = hoveredNode ?? selectedNode;
				if (!activeNode || g.hasExtremity(edge as string, activeNode)) {
					return d;
				}
				return { ...d, hidden: true };
			},
		});

		sigmaInst.on("clickNode", (payload) => {
			const node = payload.node as string;
			const kind = g.getNodeAttribute(node, "kind") as "genre" | "artist";
			const id =
				kind === "genre"
					? (g.getNodeAttribute(node, "genreId") as number)
					: (g.getNodeAttribute(node, "artistId") as number);
			onNodeTap(id, kind);
		});

		sigmaInst.on("clickStage", () => onBackgroundClick());

		sigmaInst.on("enterNode", (payload) => {
			container.style.cursor = "pointer";
			if (hoverTimer !== null) clearTimeout(hoverTimer);
			hoverTimer = setTimeout(() => {
				hoveredNode = payload.node as string;
				hoveredNeighborSet = new Set(g.neighbors(hoveredNode));
				sigmaInst?.refresh();
			}, 150);
		});

		sigmaInst.on("leaveNode", () => {
			if (hoverTimer !== null) {
				clearTimeout(hoverTimer);
				hoverTimer = null;
			}
			hoveredNode = null;
			hoveredNeighborSet = new Set();
			sigmaInst?.refresh();
			container.style.cursor = "default";
		});

		loading = false;
	}, 16);

	return () => clearTimeout(id);
});

// Selection changes must not rebuild the graph or rerun FA2 — that would
// visibly jank the layout on every click. This effect only recomputes the
// selected node's neighbor set and asks Sigma to redraw with it.
$effect(() => {
	const node = selectedNode;
	const g = graph;
	if (!sigmaInst) return;
	selectedNeighborSet =
		node && g.hasNode(node) ? new Set(g.neighbors(node)) : new Set();
	if (!hoveredNode) sigmaInst.refresh();
});

// Same reasoning as the selection effect above: dragging the time-window
// slider must only re-run the reducers (which read `activeWindow` directly)
// and redraw, never rebuild the graph or re-run FA2/Louvain.
$effect(() => {
	void activeWindow;
	if (!sigmaInst) return;
	sigmaInst.refresh();
});

// Same reasoning again: toggling "Show artists" only needs the reducers
// (which read `showArtists` directly) to re-run and redraw. The graph
// still contains every artist node/edge underneath — only their visibility
// changes — so there's nothing here to rebuild or re-layout.
$effect(() => {
	void showArtists;
	if (!sigmaInst) return;
	sigmaInst.refresh();
});

// Pans/zooms the camera to a node, used by the galaxy page's search box to
// jump straight to a match rather than requiring the user to hunt visually.
// Camera x/y must be in Sigma's normalized "framed graph" space ([0,1] range
// after Sigma fits the graph's bounding box), NOT the raw graphology x/y
// attributes FA2 assigns (which are in arbitrary graph-layout units) —
// animating the camera to raw coordinates points it at empty space and
// leaves the canvas blank. getNodeDisplayData() returns the node's position
// already converted to that normalized space.
export function focusNode(nodeId: string): void {
	if (!sigmaInst) return;
	const data = sigmaInst.getNodeDisplayData(nodeId);
	if (!data) return;
	sigmaInst
		.getCamera()
		.animate({ x: data.x, y: data.y, ratio: 0.15 }, { duration: 500 });
}
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
		flex: 1;
		min-width: 0;
		height: 100%;
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
		background: var(--bg);
	}
</style>
