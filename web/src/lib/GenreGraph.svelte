<script lang="ts">
import type Graph from "graphology";
import type Sigma from "sigma";
import { onMount } from "svelte";
import {
	cullLabelsByGrid,
	glyphSize,
	isolatedNodePosition,
	kindGlyph,
	rangesOverlap,
	shouldDrawLabel,
	type TimeRange,
} from "./graph";

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

// Sigma and its helpers are loaded dynamically so none of it is bundled into
// (or executed during) SSR; the `import type` above is erased at build time.
type SigmaModule = typeof import("sigma");
type SigmaRenderingModule = typeof import("sigma/rendering");
type NodeBorderModule = typeof import("@sigma/node-border");
type FA2Module = typeof import("graphology-layout-forceatlas2");

// Label rendering. Sigma 4's native labels are monochrome SDF text, so the
// kind glyph (color emoji) and outlined name labels are drawn on our own
// overlay canvas — see drawOverlay().
const LABEL_FONT = "'IBM Plex Sans', sans-serif";
const LABEL_SIZE = 12;
const LABEL_COLOR = "#e2e2f0";
const LABEL_RENDERED_SIZE_THRESHOLD = 8;
// Sigma 3 defaults (labelGridCellSize / labelDensity) that kept labels from piling up.
const LABEL_GRID_CELL_SIZE = 100;
const LABEL_DENSITY = 1;
const DIMMED_COLOR = "#1e1e28";
const HIGHLIGHTED_EDGE_OPACITY = 0.7;
const EDGE_CURVATURE = 0.25;

let sigmaMod = $state<SigmaModule | null>(null);
let renderingMod = $state<SigmaRenderingModule | null>(null);
let nodeBorderMod = $state<NodeBorderModule | null>(null);
let fa2Mod = $state<FA2Module | null>(null);
let sigmaInst: Sigma | null = null;
let containerEl: HTMLDivElement | undefined;
let overlayEl: HTMLCanvasElement | undefined;
let loading = $state(true);
let redrawOverlay: (() => void) | null = null;

// Hover and selection state — closed over by the Sigma reducers/event
// handlers below. Hover always takes visual priority over selection; when
// hover ends, the selection's highlight (if any) reappears.
let hoveredNode: string | null = null;
let hoveredNeighborSet = new Set<string>();
let selectedNeighborSet = new Set<string>();

onMount(() => {
	Promise.all([
		import("sigma"),
		import("sigma/rendering"),
		import("@sigma/node-border"),
		import("graphology-layout-forceatlas2"),
	]).then(([s, r, nb, f]) => {
		sigmaMod = s;
		renderingMod = r;
		nodeBorderMod = nb;
		fa2Mod = f;
	});

	// Keep the overlay's backing store in step with the container.
	const resizeObserver = new ResizeObserver(() => redrawOverlay?.());
	if (containerEl) resizeObserver.observe(containerEl);

	return () => {
		resizeObserver.disconnect();
		sigmaInst?.kill();
	};
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

interface OverlayItem {
	x: number;
	y: number;
	size: number;
	label: string;
	kind: "genre" | "artist";
}

function drawGlyph(ctx: CanvasRenderingContext2D, item: OverlayItem): void {
	ctx.save();
	ctx.font = `${glyphSize(item.size)}px sans-serif`;
	ctx.textAlign = "center";
	ctx.textBaseline = "middle";
	ctx.fillText(kindGlyph(item.kind), item.x, item.y);
	ctx.restore();
}

function drawName(
	ctx: CanvasRenderingContext2D,
	item: OverlayItem,
	outlined: boolean,
): void {
	ctx.save();
	ctx.textAlign = "left";
	ctx.textBaseline = "alphabetic";
	ctx.font = `normal ${LABEL_SIZE}px ${LABEL_FONT}`;
	const labelX = item.x + item.size + 3;
	const labelY = item.y + LABEL_SIZE / 3;
	if (outlined) {
		ctx.lineJoin = "round";
		ctx.lineWidth = 3;
		ctx.strokeStyle = "rgba(0, 0, 0, 0.95)";
		ctx.fillStyle = "#ffffff";
		ctx.strokeText(item.label, labelX, labelY);
	} else {
		ctx.fillStyle = LABEL_COLOR;
	}
	ctx.fillText(item.label, labelX, labelY);
	ctx.restore();
}

// Same pill shape as Sigma 3's default hover but with a dark background so
// the label text is readable against `--surface`.
function drawHoverPill(ctx: CanvasRenderingContext2D, item: OverlayItem): void {
	const PADDING = 2;
	ctx.save();
	ctx.font = `normal ${LABEL_SIZE}px ${LABEL_FONT}`;
	ctx.fillStyle = "#131318";
	ctx.shadowOffsetX = 0;
	ctx.shadowOffsetY = 0;
	ctx.shadowBlur = 8;
	ctx.shadowColor = "#000";

	const { x, y, size } = item;
	const textWidth = ctx.measureText(item.label).width;
	const boxWidth = Math.round(textWidth + 5);
	const boxHeight = Math.round(LABEL_SIZE + 2 * PADDING);
	const radius = Math.max(size, LABEL_SIZE / 2) + PADDING;
	const angleRadian = Math.asin(boxHeight / 2 / radius);
	const xDeltaCoord = Math.sqrt(Math.abs(radius ** 2 - (boxHeight / 2) ** 2));

	ctx.beginPath();
	ctx.moveTo(x + xDeltaCoord, y + boxHeight / 2);
	ctx.lineTo(x + radius + boxWidth, y + boxHeight / 2);
	ctx.lineTo(x + radius + boxWidth, y - boxHeight / 2);
	ctx.lineTo(x + xDeltaCoord, y - boxHeight / 2);
	ctx.arc(x, y, radius, angleRadian, -angleRadian);
	ctx.closePath();
	ctx.fill();
	ctx.restore();
}

$effect(() => {
	const g = graph;
	const sigmaModule = sigmaMod;
	const rendering = renderingMod;
	const nodeBorder = nodeBorderMod;
	const fa2 = fa2Mod;
	const container = containerEl;
	if (!sigmaModule || !rendering || !nodeBorder || !fa2 || !container) return;

	loading = true;

	const Sigma = sigmaModule.default;

	const id = setTimeout(() => {
		initCommunityLayout(g);
		fa2.default.assign(g, {
			iterations: 500,
			settings: {
				...fa2.default.inferSettings(g),
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
		const sigma = new Sigma(g, container, {
			primitives: {
				nodes: {
					shapes: [rendering.sdfCircle()],
					// 3px black outer ring, community color fill.
					layers: [
						nodeBorder.layerBorder({
							borders: [
								{ size: 3, mode: "pixels", color: "#000000" },
								{ fill: true, color: { attribute: "color" } },
							],
						}),
					],
				},
				edges: {
					paths: [rendering.pathCurved()],
					// Same bow as @sigma/edge-curve's default. pathCurved() ignores
					// its `curvature` option; the amount is a per-edge variable
					// whose default (0) is a straight line.
					variables: {
						curvature: { type: "number", default: EDGE_CURVATURE },
					},
				},
			},
			styles: {
				nodes: [
					sigmaModule.DEFAULT_STYLES.nodes,
					// No grow-on-hover and no backdrop: the hover pill is ours.
					{
						size: { attribute: "size" },
						color: { attribute: "color", defaultValue: "#7c6af7" },
						backdropVisibility: "hidden",
					},
				],
				edges: [
					sigmaModule.DEFAULT_STYLES.edges,
					{
						path: "curved",
						color: { attribute: "color", defaultValue: "#6060a0" },
						opacity: { attribute: "opacity", defaultValue: 1 },
					},
				],
			},
			settings: {
				renderLabels: false,
				renderEdgeLabels: false,
				// v3 sizes were screen pixels; v4 defaults to graph units.
				itemSizesReference: "screen",
				// Fit the camera to the graph once. The default re-fits on every
				// refresh, so hovering (which hides non-incident edges) zoomed the
				// view to whatever stayed visible.
				autoRescale: "once",
				minCameraRatio: 0.05,
				maxCameraRatio: 8,
				stagePadding: 40,
			},
			// Dim nodes/edges outside the active (hovered, or else selected) node's
			// neighborhood. Hover takes priority over selection while it's active.
			// A node/edge outside the active time window, or an artist node/edge
			// while artists are toggled off, is hidden outright, ahead of and
			// regardless of hover/selection dimming.
			nodeReducer: (node, _display, attrs) => {
				if (!showArtists && attrs.kind === "artist") {
					return { visibility: "hidden" };
				}
				if (activeWindow) {
					const ranges: TimeRange[] = attrs.ranges ?? [];
					if (!rangesOverlap(ranges, activeWindow[0], activeWindow[1])) {
						return { visibility: "hidden" };
					}
				}
				const activeNode = hoveredNode ?? selectedNode;
				const activeNeighbors = hoveredNode
					? hoveredNeighborSet
					: selectedNeighborSet;
				if (!activeNode || node === activeNode || activeNeighbors.has(node)) {
					return {};
				}
				return { color: DIMMED_COLOR, label: "" };
			},
			edgeReducer: (edge, _display, attrs) => {
				if (!showArtists && attrs.kind !== "genre-genre") {
					return { visibility: "hidden" };
				}
				if (activeWindow) {
					const ranges: TimeRange[] = attrs.ranges ?? [];
					if (!rangesOverlap(ranges, activeWindow[0], activeWindow[1])) {
						return { visibility: "hidden" };
					}
				}
				const activeNode = hoveredNode ?? selectedNode;
				if (!activeNode) return {};
				// A focused node's own edges are emphasized: at their resting opacity
				// they read as faint, which is right for the whole web but not for
				// the handful the user is actually looking at.
				if (g.hasExtremity(edge, activeNode)) {
					return { opacity: HIGHLIGHTED_EDGE_OPACITY };
				}
				return { visibility: "hidden" };
			},
		});
		sigmaInst = sigma;

		// Overlay canvas: kind glyph centered on each labelled node, plus the
		// name as solid text over a near-opaque black stroke (maximum contrast
		// against the dense web of edges). The hovered node gets a dark pill
		// so its label stays readable. Redrawn after every sigma render.
		function drawOverlay(): void {
			const canvas = overlayEl;
			const ctx = canvas?.getContext("2d");
			if (!canvas || !ctx || !container) return;
			const { width, height } = container.getBoundingClientRect();
			const dpr = window.devicePixelRatio || 1;
			if (
				canvas.width !== Math.round(width * dpr) ||
				canvas.height !== Math.round(height * dpr)
			) {
				canvas.width = Math.round(width * dpr);
				canvas.height = Math.round(height * dpr);
			}
			ctx.setTransform(dpr, 0, 0, dpr, 0, 0);
			ctx.clearRect(0, 0, width, height);

			let hoveredDraw: OverlayItem | null = null;

			const candidates: OverlayItem[] = [];
			g.forEachNode((node, attrs) => {
				const display = sigma.getNodeDisplayData(node);
				if (!display || display.visibility === "hidden") return;
				if (!display.label) return; // dimmed by the reducer
				const size = sigma.scaleSize(display.size);
				const isHovered = node === hoveredNode;
				if (!shouldDrawLabel(size, LABEL_RENDERED_SIZE_THRESHOLD, isHovered))
					return;
				const { x, y } = sigma.graphToViewport({ x: attrs.x, y: attrs.y });
				if (x < -100 || y < -20 || x > width + 100 || y > height + 20) return;
				const item: OverlayItem = {
					x,
					y,
					size,
					label: display.label,
					kind: attrs.kind as "genre" | "artist",
				};
				if (isHovered) {
					hoveredDraw = item;
					return;
				}
				candidates.push(item);
			});

			// Smallest first so the largest nodes' glyphs paint on top.
			const visible = cullLabelsByGrid(
				candidates,
				LABEL_GRID_CELL_SIZE,
				LABEL_DENSITY,
			).reverse();
			for (const item of visible) {
				drawGlyph(ctx, item);
				drawName(ctx, item, true);
			}

			if (hoveredDraw) {
				drawHoverPill(ctx, hoveredDraw);
				drawGlyph(ctx, hoveredDraw);
				drawName(ctx, hoveredDraw, false);
			}
		}

		redrawOverlay = drawOverlay;
		sigma.on("afterRender", drawOverlay);

		sigma.on("clickNode", (payload) => {
			const node = payload.node;
			const kind: "genre" | "artist" = g.getNodeAttribute(node, "kind");
			const id =
				kind === "genre"
					? g.getNodeAttribute(node, "genreId")
					: g.getNodeAttribute(node, "artistId");
			onNodeTap(id, kind);
		});

		sigma.on("clickStage", () => onBackgroundClick());

		sigma.on("enterNode", (payload) => {
			container.style.cursor = "pointer";
			if (hoverTimer !== null) clearTimeout(hoverTimer);
			hoverTimer = setTimeout(() => {
				hoveredNode = payload.node;
				hoveredNeighborSet = new Set(g.neighbors(hoveredNode));
				sigmaInst?.refresh();
			}, 150);
		});

		sigma.on("leaveNode", () => {
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
	<canvas class="graph-overlay" bind:this={overlayEl}></canvas>
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

	.graph-overlay {
		position: absolute;
		inset: 0;
		width: 100%;
		height: 100%;
		pointer-events: none;
	}

	.graph-container {
		width: 100%;
		height: 100%;
		background: var(--bg);
	}
</style>
