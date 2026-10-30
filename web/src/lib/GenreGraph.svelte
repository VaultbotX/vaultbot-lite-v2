<script lang="ts">
import Graph from "graphology";
import type Sigma from "sigma";
import { onMount } from "svelte";
import {
	computeNeighborhoodCameraState,
	createSeededRng,
	cullLabelsByGrid,
	cullOverlappingLabels,
	type ExternalPull,
	kindGlyph,
	layoutConstellationCenters,
	optimalConstellationRotation,
	placeSatelliteComponents,
	separateCircles,
	shouldDrawLabel,
	type WindowVisibility,
} from "./graph";

let {
	graph,
	selectedNode,
	windowVisibility,
	showArtists,
	onNodeTap,
	onBackgroundClick,
}: {
	graph: Graph;
	selectedNode: string | null;
	windowVisibility: WindowVisibility | null;
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
type LayoutModule = typeof import("graphology-layout");

interface CustomNodeState {
	isDimmed: boolean;
	isFilterHidden: boolean;
}

interface CustomEdgeState {
	isDimmed: boolean;
	isFilterHidden: boolean;
}

type GalaxySigma = Sigma<
	Record<string, unknown>,
	Record<string, unknown>,
	Record<string, unknown>,
	CustomNodeState,
	CustomEdgeState
>;

// Label rendering. Sigma 4's native labels are monochrome SDF text, so
// outlined name labels and hover pills are drawn on our own overlay canvas.
const LABEL_FONT = "'IBM Plex Sans', sans-serif";
const LABEL_SIZE = 12;
const LABEL_COLOR = "#e2e2f0";
const LABEL_RENDERED_SIZE_THRESHOLD = 6.2;
// Spatial label grid cell size & density keeping constellation labels readable.
const LABEL_GRID_CELL_SIZE = 120;
const LABEL_DENSITY = 1;
const DIMMED_COLOR = "#1a1a24";
const HIGHLIGHTED_EDGE_COLOR = "#9d8df9";
const EDGE_CURVATURE = 0.2;
const CORE_COMPONENT_MIN_NODES = 10;
const CIRCLEPACK_SEED = 1337;
const LOCAL_FA2_ITERATIONS = 48;
const GALAXY_ASPECT_X = 1.85;
const CONSTELLATION_CHANNEL_GAP = 54;

let sigmaMod = $state<SigmaModule | null>(null);
let renderingMod = $state<SigmaRenderingModule | null>(null);
let nodeBorderMod = $state<NodeBorderModule | null>(null);
let fa2Mod = $state<FA2Module | null>(null);
let layoutMod = $state<LayoutModule | null>(null);
let sigmaInst: GalaxySigma | null = null;
let containerEl: HTMLDivElement | undefined;
let overlayEl: HTMLCanvasElement | undefined;
let loading = $state(true);
let redrawOverlay: (() => void) | null = null;

// Hover and selection state.
let hoveredNode = $state<string | null>(null);
let activeHighlightNode: string | null = null;
let currentHighlightedNodes: string[] = [];
let currentHighlightedEdges: string[] = [];
let currentDimmedNodes: string[] = [];
let currentDimmedEdges: string[] = [];
let applyFilters: (() => void) | null = null;
let updateActiveHighlight: ((target: string | null) => void) | null = null;

onMount(() => {
	Promise.all([
		import("sigma"),
		import("sigma/rendering"),
		import("@sigma/node-border"),
		import("graphology-layout-forceatlas2"),
		import("graphology-layout"),
	]).then(([s, r, nb, f, l]) => {
		sigmaMod = s;
		renderingMod = r;
		nodeBorderMod = nb;
		fa2Mod = f;
		layoutMod = l;
	});

	// Keep both Sigma's WebGL viewport and the 2D overlay canvas in step with
	// container size changes (e.g. when the detail drawer opens or closes).
	const resizeObserver = new ResizeObserver(() => {
		if (!sigmaInst) return;
		sigmaInst.resize();
		sigmaInst.scheduleRefresh({ skipIndexation: true });
		redrawOverlay?.();
	});
	if (containerEl) resizeObserver.observe(containerEl);

	return () => {
		resizeObserver.disconnect();
		sigmaInst?.kill();
	};
});

interface ConstellationCluster {
	commId: number;
	nodes: { id: string; x: number; y: number; personalR: number }[];
	radius: number;
}

/**
 * Two-level constellation layout:
 * 1. Lays out each community locally (hub-centered circlepack seed -> local FA2
 *    with strong gravity -> uniform star-density RMS normalization -> circle
 *    overlap separation) so major genre hubs anchor the heart of their nebula.
 * 2. Positions constellation centers across a widescreen galactic plane with
 *    guaranteed dark space channels (`CONSTELLATION_CHANNEL_GAP`), then rotates
 *    each constellation via Procrustes alignment so bridge genres face their
 *    neighboring constellations.
 * 3. Places disconnected islands along a subtle outer elliptical belt.
 */
function assignGalaxyLayout(
	g: Graph,
	circlepack: LayoutModule["circlepack"],
	fa2: FA2Module["default"],
): void {
	const visited = new Set<string>();
	const coreNodeIds = new Set<string>();
	const satellites: string[][] = [];

	g.forEachNode((start) => {
		if (visited.has(start)) return;
		const queue = [start];
		visited.add(start);
		for (let i = 0; i < queue.length; i++) {
			const u = queue[i];
			g.forEachNeighbor(u, (v) => {
				if (!visited.has(v)) {
					visited.add(v);
					queue.push(v);
				}
			});
		}
		if (queue.length >= CORE_COMPONENT_MIN_NODES) {
			for (const id of queue) coreNodeIds.add(id);
		} else {
			satellites.push(queue);
		}
	});

	const coreCoords: { x: number; y: number }[] = [];
	if (coreNodeIds.size > 0) {
		const byComm = new Map<number, string[]>();
		for (const id of coreNodeIds) {
			const comm = g.getNodeAttribute(id, "community") as number;
			let list = byComm.get(comm);
			if (!list) {
				list = [];
				byComm.set(comm, list);
			}
			list.push(id);
		}

		const constellations: ConstellationCluster[] = [];
		const rng = createSeededRng(CIRCLEPACK_SEED);

		for (const [commId, memberIds] of [...byComm.entries()].sort(
			(a, b) => b[1].length - a[1].length || a[0] - b[0],
		)) {
			const sortedMembers = [...memberIds].sort(
				(a, b) =>
					(g.getNodeAttribute(b, "size") as number) -
					(g.getNodeAttribute(a, "size") as number),
			);
			const memberSet = new Set(sortedMembers);
			const sub = new Graph({ type: "undirected", multi: false });
			for (const id of sortedMembers) {
				sub.addNode(id, { size: g.getNodeAttribute(id, "size") as number });
			}

			for (const id of sortedMembers) {
				let hasLocalBackbone = false;
				let bestFallback: string | null = null;
				let bestFallbackW = -1;
				g.forEachEdge(id, (_e, eAttrs, src, tgt) => {
					const other = src === id ? tgt : src;
					if (!memberSet.has(other)) return;
					const nw = (eAttrs.normWeight as number | undefined) ?? 0.5;
					if (eAttrs.backbone) {
						hasLocalBackbone = true;
						if (!sub.hasEdge(id, other)) {
							sub.addEdge(id, other, { weight: nw });
						}
					} else if (nw > bestFallbackW) {
						bestFallbackW = nw;
						bestFallback = other;
					}
				});
				if (
					!hasLocalBackbone &&
					bestFallback !== null &&
					!sub.hasEdge(id, bestFallback)
				) {
					sub.addEdge(id, bestFallback, { weight: bestFallbackW });
				}
			}

			circlepack.assign(sub, { scale: 2.0, rng });

			if (sub.order > 1) {
				fa2.assign(sub, {
					iterations: LOCAL_FA2_ITERATIONS,
					settings: {
						barnesHutOptimize: false,
						linLogMode: false,
						outboundAttractionDistribution: false,
						strongGravityMode: true,
						gravity: 1.6,
						scalingRatio: 10,
						slowDown: 2.0,
						adjustSizes: false,
					},
					getEdgeWeight: "weight",
				});
			}

			const localNodes: {
				id: string;
				x: number;
				y: number;
				personalR: number;
			}[] = [];
			let cx = 0;
			let cy = 0;
			let areaSum = 0;
			sub.forEachNode((id, attrs) => {
				const size = attrs.size as number;
				const personalR = size * 1.45 + 2.4;
				areaSum += personalR * personalR;
				cx += attrs.x as number;
				cy += attrs.y as number;
				localNodes.push({
					id,
					x: attrs.x as number,
					y: attrs.y as number,
					personalR,
				});
			});
			cx /= localNodes.length;
			cy /= localNodes.length;

			let sumSq = 0;
			for (const n of localNodes) {
				n.x -= cx;
				n.y -= cy;
				sumSq += n.x * n.x + n.y * n.y;
			}
			const rms = Math.sqrt(sumSq / localNodes.length) || 1;
			const targetRms = 1.05 * Math.sqrt(areaSum);
			const scale = targetRms / rms;
			for (const n of localNodes) {
				n.x *= scale;
				n.y *= scale;
			}

			separateCircles(
				localNodes,
				localNodes.map((n) => n.personalR),
				16,
			);

			let fcx = 0;
			let fcy = 0;
			for (const n of localNodes) {
				fcx += n.x;
				fcy += n.y;
			}
			fcx /= localNodes.length;
			fcy /= localNodes.length;

			let maxR = 12;
			for (const n of localNodes) {
				n.x -= fcx;
				n.y -= fcy;
				const r = Math.hypot(n.x, n.y) + n.personalR;
				if (r > maxR) maxR = r;
			}

			constellations.push({ commId, nodes: localNodes, radius: maxR });
		}

		const commIndex = new Map<number, number>(
			constellations.map((c, idx) => [c.commId, idx]),
		);
		const numC = constellations.length;
		const affinities = Array.from({ length: numC }, () =>
			new Array<number>(numC).fill(0),
		);

		g.forEachEdge((_e, eAttrs, src, tgt) => {
			if (!coreNodeIds.has(src) || !coreNodeIds.has(tgt)) return;
			const cs = commIndex.get(g.getNodeAttribute(src, "community") as number);
			const ct = commIndex.get(g.getNodeAttribute(tgt, "community") as number);
			if (cs === undefined || ct === undefined || cs === ct) return;
			const w = (eAttrs.normWeight as number | undefined) ?? 0.5;
			affinities[cs][ct] += w;
			affinities[ct][cs] += w;
		});

		const centers = layoutConstellationCenters(
			constellations.map((c) => c.radius),
			affinities,
			{ aspectX: GALAXY_ASPECT_X, channelGap: CONSTELLATION_CHANNEL_GAP },
		);

		const pullsByComm = Array.from({ length: numC }, (): ExternalPull[] => []);
		g.forEachEdge((_e, eAttrs, src, tgt) => {
			if (!coreNodeIds.has(src) || !coreNodeIds.has(tgt)) return;
			const cs = commIndex.get(g.getNodeAttribute(src, "community") as number);
			const ct = commIndex.get(g.getNodeAttribute(tgt, "community") as number);
			if (cs === undefined || ct === undefined || cs === ct) return;
			const w = (eAttrs.normWeight as number | undefined) ?? 0.5;
			const dx = centers[ct].x - centers[cs].x;
			const dy = centers[ct].y - centers[cs].y;
			pullsByComm[cs].push({ nodeId: src, dirX: dx, dirY: dy, weight: w });
			pullsByComm[ct].push({ nodeId: tgt, dirX: -dx, dirY: -dy, weight: w });
		});

		for (let i = 0; i < numC; i++) {
			const c = constellations[i];
			const center = centers[i];
			const theta = optimalConstellationRotation(c.nodes, pullsByComm[i]);
			const cos = Math.cos(theta);
			const sin = Math.sin(theta);
			for (const n of c.nodes) {
				const worldX = center.x + n.x * cos - n.y * sin;
				const worldY = center.y + n.x * sin + n.y * cos;
				g.setNodeAttribute(n.id, "x", worldX);
				g.setNodeAttribute(n.id, "y", worldY);
				coreCoords.push({ x: worldX, y: worldY });
			}
		}
	}

	const satelliteCoords = placeSatelliteComponents(coreCoords, satellites);
	for (const [id, pos] of satelliteCoords) {
		g.setNodeAttribute(id, "x", pos.x);
		g.setNodeAttribute(id, "y", pos.y);
	}
}

interface OverlayItem {
	node: string;
	x: number;
	y: number;
	size: number;
	color: string;
	label: string;
	rank: number | undefined;
	kind: "genre" | "artist";
}

function pillText(item: OverlayItem): string {
	const rankSuffix = item.rank !== undefined ? ` · #${item.rank}` : "";
	return `${kindGlyph(item.kind)} ${item.label}${rankSuffix}`;
}

function drawName(
	ctx: CanvasRenderingContext2D,
	item: OverlayItem,
	outlined: boolean,
	text = item.label,
): void {
	ctx.save();
	ctx.textAlign = "left";
	ctx.textBaseline = "alphabetic";
	const weight = item.kind === "genre" ? "500" : "normal";
	ctx.font = `${weight} ${LABEL_SIZE}px ${LABEL_FONT}`;
	const offset = outlined
		? item.size + 4
		: Math.max(item.size, LABEL_SIZE / 2) + 6;
	const labelX = item.x + offset;
	const labelY = item.y + LABEL_SIZE / 3;
	if (outlined) {
		ctx.lineJoin = "round";
		ctx.lineWidth = 3;
		ctx.strokeStyle = "rgba(12, 12, 16, 0.92)";
		ctx.fillStyle = item.kind === "genre" ? "#f0f0fa" : "#c8c8dc";
		ctx.strokeText(text, labelX, labelY);
	} else {
		ctx.fillStyle = LABEL_COLOR;
	}
	ctx.fillText(text, labelX, labelY);
	ctx.restore();
}

// Dark pill badge with shadow + border, repainting the node's vivid colored
// circle and crisp light halo ring inside the left cap so the focused node
// never gets covered by the pill backdrop.
function drawHoverPill(
	ctx: CanvasRenderingContext2D,
	item: OverlayItem,
	text: string,
): void {
	const PADDING = 4;
	ctx.save();
	const weight = item.kind === "genre" ? "500" : "normal";
	ctx.font = `${weight} ${LABEL_SIZE}px ${LABEL_FONT}`;
	ctx.fillStyle = "#131318";
	ctx.strokeStyle = "#2d2d44";
	ctx.lineWidth = 1;
	ctx.shadowOffsetX = 0;
	ctx.shadowOffsetY = 2;
	ctx.shadowBlur = 10;
	ctx.shadowColor = "rgba(0, 0, 0, 0.85)";

	const { x, y, size } = item;
	const discRadius = Math.max(size, 4);
	const textWidth = ctx.measureText(text).width;
	const boxWidth = Math.round(textWidth + 12);
	const boxHeight = Math.round(LABEL_SIZE + 2 * PADDING);
	const radius = Math.max(discRadius, LABEL_SIZE / 2) + PADDING;
	const angleRadian = Math.asin(Math.min(1, boxHeight / 2 / radius));
	const xDeltaCoord = Math.sqrt(Math.abs(radius ** 2 - (boxHeight / 2) ** 2));

	ctx.beginPath();
	ctx.moveTo(x + xDeltaCoord, y + boxHeight / 2);
	ctx.lineTo(x + radius + boxWidth, y + boxHeight / 2);
	ctx.lineTo(x + radius + boxWidth, y - boxHeight / 2);
	ctx.lineTo(x + xDeltaCoord, y - boxHeight / 2);
	ctx.arc(x, y, radius, angleRadian, -angleRadian);
	ctx.closePath();
	ctx.fill();
	ctx.stroke();
	ctx.restore();

	ctx.save();
	ctx.beginPath();
	ctx.arc(x, y, discRadius, 0, Math.PI * 2);
	ctx.fillStyle = item.color;
	ctx.fill();
	ctx.lineWidth = 1.6;
	ctx.strokeStyle = "#f0f0fa";
	ctx.stroke();
	ctx.restore();
}

function isEdgeActiveInFilter(
	attrs: Record<string, unknown>,
	artistsVisible: boolean,
	winVis: WindowVisibility | null,
): boolean {
	if (!artistsVisible && attrs.kind !== "genre-genre") return false;
	if (!winVis) return true;
	if (attrs.kind === "genre-artist") {
		return winVis.artistIds.has(attrs.artistId as number);
	}
	if (attrs.kind === "genre-genre") {
		return winVis.genreGenrePairs.has(attrs.pairKey as string);
	}
	return winVis.artistArtistPairs.has(attrs.pairKey as string);
}

function activeNeighborsOf(
	g: Graph,
	nodeId: string,
	artistsVisible: boolean,
	winVis: WindowVisibility | null,
): Set<string> {
	const neighbors = new Set<string>();
	if (!g.hasNode(nodeId)) return neighbors;
	g.forEachEdge(nodeId, (_edge, attrs, src, tgt) => {
		if (!isEdgeActiveInFilter(attrs, artistsVisible, winVis)) return;
		neighbors.add(src === nodeId ? tgt : src);
	});
	return neighbors;
}

function activeIncidentEdgesOf(
	g: Graph,
	nodeId: string,
	artistsVisible: boolean,
	winVis: WindowVisibility | null,
): string[] {
	const edges: string[] = [];
	if (!g.hasNode(nodeId)) return edges;
	g.forEachEdge(nodeId, (edge, attrs) => {
		if (!isEdgeActiveInFilter(attrs, artistsVisible, winVis)) return;
		edges.push(edge);
	});
	return edges;
}

$effect(() => {
	const g = graph;
	const sigmaModule = sigmaMod;
	const rendering = renderingMod;
	const nodeBorder = nodeBorderMod;
	const fa2 = fa2Mod;
	const layout = layoutMod;
	const container = containerEl;
	if (
		!sigmaModule ||
		!rendering ||
		!nodeBorder ||
		!fa2 ||
		!layout ||
		!container
	)
		return;

	loading = true;

	const Sigma = sigmaModule.default;

	const id = setTimeout(() => {
		assignGalaxyLayout(g, layout.circlepack, fa2.default);

		let hoverTimer: ReturnType<typeof setTimeout> | null = null;
		sigmaInst?.kill();
		const sigma = new Sigma(g, container, {
			customNodeState: {
				isDimmed: false,
				isFilterHidden: false,
			},
			customEdgeState: {
				isDimmed: false,
				isFilterHidden: false,
			},
			primitives: {
				// Interleave edges and nodes so highlighted edges (topEdges) draw
				// above dimmed nodes, while highlighted nodes (topNodes) draw above
				// those edges.
				depthLayers: ["edges", "nodes", "topEdges", "topNodes"],
				nodes: {
					shapes: [rendering.sdfCircle()],
					// Crisp 1.2px `--bg` separator ring so small nodes keep vivid fills.
					layers: [
						nodeBorder.layerBorder({
							borders: [
								{ size: 1.2, mode: "pixels", color: "#0c0c10" },
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
					{
						size: { attribute: "size" },
						color: { attribute: "color", defaultValue: "#7c6af7" },
						label: { attribute: "label" },
						labelColor: LABEL_COLOR,
						labelSize: LABEL_SIZE,
						labelFont: LABEL_FONT,
						backdropVisibility: "hidden",
						depth: "nodes",
						visibility: {
							whenState: "isFilterHidden",
							then: "hidden",
							else: "visible",
						},
					},
					{
						whenState: "isDimmed",
						then: {
							color: DIMMED_COLOR,
							label: "",
							depth: "nodes",
						},
					},
					{
						whenState: "isHighlighted",
						then: {
							color: { attribute: "color" },
							depth: "topNodes",
						},
					},
					{
						whenState: "isHovered",
						then: {
							color: { attribute: "color" },
							depth: "topNodes",
						},
					},
				],
				edges: [
					sigmaModule.DEFAULT_STYLES.edges,
					{
						path: "curved",
						color: { attribute: "color", defaultValue: "#6060a0" },
						opacity: { attribute: "opacity", defaultValue: 1 },
						size: { attribute: "size", defaultValue: 0.5 },
						depth: "edges",
						visibility: {
							whenState: "isFilterHidden",
							then: "hidden",
							else: {
								whenData: "backbone",
								then: "visible",
								else: "hidden",
							},
						},
					},
					{
						whenState: "isDimmed",
						then: {
							visibility: "hidden",
						},
					},
					{
						whenState: "isHighlighted",
						then: {
							color: HIGHLIGHTED_EDGE_COLOR,
							opacity: { attribute: "highlightedOpacity", defaultValue: 0.7 },
							size: { attribute: "highlightedSize", defaultValue: 1.5 },
							depth: "topEdges",
							visibility: "visible",
						},
					},
				],
			},
			settings: {
				renderLabels: true,
				renderEdgeLabels: false,
				enableEdgeEvents: false,
				hideEdgesOnMove: true,
				labelGridCellSize: LABEL_GRID_CELL_SIZE,
				labelRenderedSizeThreshold: LABEL_RENDERED_SIZE_THRESHOLD,
				labelDensity: LABEL_DENSITY,
				// Allow hairline edges down to 0.4px instead of Sigma 4's 1.7px floor,
				// which otherwise inflates thousands of weak edges into an opaque mass.
				minEdgeThickness: 0.4,
				// v3 sizes were screen pixels; v4 defaults to graph units.
				itemSizesReference: "screen",
				// Fit the camera to the graph once. The default re-fits on every
				// refresh, so hovering (which hides non-incident edges) zoomed the
				// view to whatever stayed visible.
				autoRescale: "once",
				minCameraRatio: 0.02,
				maxCameraRatio: 8,
				stagePadding: 40,
			},
		});
		sigmaInst = sigma;

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

			const sizeScale = sigma.scaleSize(1);

			const drawNodePill = (nodeId: string) => {
				const display = sigma.getNodeDisplayData(nodeId);
				if (!display || display.visibility === "hidden") return;
				const { x, y } = sigma.framedGraphToViewport(display);
				if (x < -100 || y < -20 || x > width + 100 || y > height + 20) return;
				const item: OverlayItem = {
					node: nodeId,
					x,
					y,
					size: display.size * sizeScale,
					color:
						(g.getNodeAttribute(nodeId, "color") as string | undefined) ??
						"#7c6af7",
					label: display.label ?? "",
					rank: g.getNodeAttribute(nodeId, "rank") as number | undefined,
					kind: g.getNodeAttribute(nodeId, "kind") as "genre" | "artist",
				};
				const text = pillText(item);
				drawHoverPill(ctx, item, text);
				drawName(ctx, item, false, text);
			};

			if (
				selectedNode &&
				selectedNode !== hoveredNode &&
				g.hasNode(selectedNode)
			) {
				drawNodePill(selectedNode);
			}

			if (hoveredNode && g.hasNode(hoveredNode)) {
				drawNodePill(hoveredNode);
			}
		}

		redrawOverlay = drawOverlay;
		sigma.on("afterRender", drawOverlay);
		drawOverlay();
		if (selectedNode && g.hasNode(selectedNode)) {
			focusNode(selectedNode, 0);
		}

		applyFilters = () => {
			if (!sigmaInst) return;
			const nodesToHide: string[] = [];
			const nodesToShow: string[] = [];

			g.forEachNode((node, attrs) => {
				const isArtist = attrs.kind === "artist";
				if (!showArtists && isArtist) {
					nodesToHide.push(node);
					return;
				}
				if (windowVisibility) {
					const inWindow =
						attrs.kind === "genre"
							? windowVisibility.genreIds.has(attrs.genreId as number)
							: windowVisibility.artistIds.has(attrs.artistId as number);
					if (!inWindow) {
						nodesToHide.push(node);
						return;
					}
				}
				nodesToShow.push(node);
			});

			sigmaInst.setNodesState(nodesToHide, { isFilterHidden: true });
			sigmaInst.setNodesState(nodesToShow, { isFilterHidden: false });

			const edgesToHide: string[] = [];
			const edgesToShow: string[] = [];
			g.forEachEdge((edge, attrs) => {
				if (!isEdgeActiveInFilter(attrs, showArtists, windowVisibility)) {
					edgesToHide.push(edge);
				} else {
					edgesToShow.push(edge);
				}
			});
			sigmaInst.setEdgesState(edgesToHide, { isFilterHidden: true });
			sigmaInst.setEdgesState(edgesToShow, { isFilterHidden: false });
		};

		updateActiveHighlight = (targetNode: string | null) => {
			if (!sigmaInst) return;
			if (targetNode === activeHighlightNode) return;

			// Clear previous states
			if (activeHighlightNode && g.hasNode(activeHighlightNode)) {
				sigmaInst.setNodeState(activeHighlightNode, {
					isHovered: false,
					isHighlighted: false,
					isDimmed: false,
				});
			}
			if (currentHighlightedNodes.length > 0) {
				sigmaInst.setNodesState(currentHighlightedNodes, {
					isHighlighted: false,
					isDimmed: false,
				});
				currentHighlightedNodes = [];
			}
			if (currentHighlightedEdges.length > 0) {
				sigmaInst.setEdgesState(currentHighlightedEdges, {
					isHighlighted: false,
					isDimmed: false,
				});
				currentHighlightedEdges = [];
			}
			if (currentDimmedNodes.length > 0) {
				sigmaInst.setNodesState(currentDimmedNodes, { isDimmed: false });
				currentDimmedNodes = [];
			}
			if (currentDimmedEdges.length > 0) {
				sigmaInst.setEdgesState(currentDimmedEdges, { isDimmed: false });
				currentDimmedEdges = [];
			}

			activeHighlightNode = targetNode;
			if (!targetNode || !g.hasNode(targetNode)) {
				redrawOverlay?.();
				return;
			}

			const neighbors = activeNeighborsOf(
				g,
				targetNode,
				showArtists,
				windowVisibility,
			);
			const incidentEdges = activeIncidentEdgesOf(
				g,
				targetNode,
				showArtists,
				windowVisibility,
			);

			currentHighlightedNodes = Array.from(neighbors);
			currentHighlightedEdges = incidentEdges;

			const nodesToDim: string[] = [];
			g.forEachNode((node) => {
				if (node !== targetNode && !neighbors.has(node)) {
					nodesToDim.push(node);
				}
			});
			currentDimmedNodes = nodesToDim;

			const edgesToDim: string[] = [];
			const incidentSet = new Set(incidentEdges);
			g.forEachEdge((edge, attrs) => {
				if (attrs.backbone && !incidentSet.has(edge)) {
					edgesToDim.push(edge);
				}
			});
			currentDimmedEdges = edgesToDim;

			sigmaInst.setNodesState(nodesToDim, { isDimmed: true });
			sigmaInst.setEdgesState(edgesToDim, { isDimmed: true });
			sigmaInst.setNodeState(targetNode, {
				isHovered: true,
				isDimmed: false,
			});
			sigmaInst.setNodesState(currentHighlightedNodes, {
				isHighlighted: true,
				isDimmed: false,
			});
			sigmaInst.setEdgesState(currentHighlightedEdges, {
				isHighlighted: true,
				isDimmed: false,
			});

			redrawOverlay?.();
		};

		applyFilters();
		updateActiveHighlight(selectedNode);

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
			hoveredNode = payload.node;
			updateActiveHighlight?.(payload.node);
		});

		sigma.on("leaveNode", () => {
			container.style.cursor = "default";
			if (hoveredNode !== null) {
				hoveredNode = null;
				updateActiveHighlight?.(selectedNode);
			}
		});

		loading = false;
	}, 16);

	return () => clearTimeout(id);
});

// Selection, time-window, and artist-toggle changes
$effect(() => {
	const node = selectedNode;
	const _show = showArtists;
	const _win = windowVisibility;
	if (!sigmaInst) return;
	applyFilters?.();
	updateActiveHighlight?.(hoveredNode ?? node);
});

// Pans and zooms the camera to frame `nodeId` together with all of its
// currently visible connected neighbors so the user can see the full web of
// connections for the selected node at once. Camera coordinates live in
// Sigma's normalized "framed graph" space (`getNodeDisplayData()`), and
// viewport width accounts for the 380px detail drawer settling open alongside
// the canvas.
export function focusNode(nodeId: string, duration = 500): void {
	if (!sigmaInst || !containerEl) return;
	const data = sigmaInst.getNodeDisplayData(nodeId);
	if (!data) return;

	const points: { x: number; y: number }[] = [{ x: data.x, y: data.y }];
	for (const neighborId of activeNeighborsOf(
		graph,
		nodeId,
		showArtists,
		windowVisibility,
	)) {
		const nData = sigmaInst.getNodeDisplayData(neighborId);
		if (nData) points.push({ x: nData.x, y: nData.y });
	}

	const rect = containerEl.getBoundingClientRect();
	const drawerEl = containerEl
		.closest(".graph-row")
		?.querySelector<HTMLElement>(".drawer");
	const currentDrawerWidth = drawerEl?.getBoundingClientRect().width ?? 0;
	const pendingDrawerDelta =
		window.innerWidth > 900 ? Math.max(0, 380 - currentDrawerWidth) : 0;

	const targetState = computeNeighborhoodCameraState({
		points,
		viewport: {
			width: Math.max(200, rect.width - pendingDrawerDelta),
			height: Math.max(200, rect.height),
		},
		graphDimensions: sigmaInst.getGraphDimensions(),
		stagePadding: sigmaInst.getStagePadding(),
	});
	if (!targetState) return;

	if (duration <= 0) {
		sigmaInst.getCamera().setState(targetState);
	} else {
		sigmaInst.getCamera().animate(targetState, { duration });
	}
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
