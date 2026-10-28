<script lang="ts">
import Graph from "graphology";
import type Sigma from "sigma";
import { onMount } from "svelte";
import {
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
		selectedNeighborSet =
			selectedNode && g.hasNode(selectedNode)
				? new Set(g.neighbors(selectedNode))
				: new Set();

		sigmaInst?.kill();
		const sigma = new Sigma(g, container, {
			primitives: {
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
				enableEdgeEvents: false,
				hideEdgesOnMove: true,
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
			// Dim nodes/edges outside the active (hovered, or else selected) node's
			// neighborhood. Hover takes priority over selection while it's active.
			// A node/edge outside the active time window, or an artist node/edge
			// while artists are toggled off, is hidden outright, ahead of and
			// regardless of hover/selection dimming.
			nodeReducer: (node, _display, attrs) => {
				if (!showArtists && attrs.kind === "artist") {
					return { visibility: "hidden" };
				}
				if (windowVisibility) {
					const inWindow =
						attrs.kind === "genre"
							? windowVisibility.genreIds.has(attrs.genreId as number)
							: windowVisibility.artistIds.has(attrs.artistId as number);
					if (!inWindow) {
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
				if (windowVisibility) {
					let inWindow = false;
					if (attrs.kind === "genre-artist") {
						inWindow = windowVisibility.artistIds.has(attrs.artistId as number);
					} else if (attrs.kind === "genre-genre") {
						inWindow = windowVisibility.genreGenrePairs.has(
							attrs.pairKey as string,
						);
					} else {
						inWindow = windowVisibility.artistArtistPairs.has(
							attrs.pairKey as string,
						);
					}
					if (!inWindow) {
						return { visibility: "hidden" };
					}
				}
				const activeNode = hoveredNode ?? selectedNode;
				if (!activeNode) {
					return attrs.backbone ? {} : { visibility: "hidden" };
				}
				// Emphasize every incident edge (including non-backbone links) when a
				// node is hovered or selected, scaling opacity by weight so a hub's
				// strongest ties stand out without 400 weak links forming a solid fan.
				if (g.hasExtremity(edge, activeNode)) {
					const nw = (attrs.normWeight as number | undefined) ?? 0.5;
					return {
						opacity: 0.14 + 0.58 * nw,
						color: HIGHLIGHTED_EDGE_COLOR,
						size: 0.45 + 1.55 * nw,
					};
				}
				return { visibility: "hidden" };
			},
		});
		sigmaInst = sigma;

		// Overlay canvas: outlined hub labels culled by a spatial grid, plus a
		// dark badge pill with kind glyph + rank for the hovered/selected node.
		let labelledNodes: Set<string> | null = null;

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
			let selectedDraw: OverlayItem | null = null;
			const sizeScale = sigma.scaleSize(1);

			const candidates: OverlayItem[] = [];
			g.forEachNode((node, attrs) => {
				const display = sigma.getNodeDisplayData(node);
				if (!display || display.visibility === "hidden") return;
				if (!display.label) return; // dimmed by the reducer
				const size = display.size * sizeScale;
				const isHovered = node === hoveredNode;
				const isSelected = node === selectedNode;
				if (
					!isSelected &&
					!shouldDrawLabel(size, LABEL_RENDERED_SIZE_THRESHOLD, isHovered)
				)
					return;
				const { x, y } = sigma.framedGraphToViewport(display);
				if (x < -100 || y < -20 || x > width + 100 || y > height + 20) return;
				const item: OverlayItem = {
					node,
					x,
					y,
					size,
					color: g.getNodeAttribute(node, "color") as string,
					label: display.label,
					rank: attrs.rank as number | undefined,
					kind: attrs.kind as "genre" | "artist",
				};
				if (isHovered) {
					hoveredDraw = item;
					return;
				}
				if (isSelected) {
					selectedDraw = item;
					return;
				}
				candidates.push(item);
			});

			// The density cull buckets by screen position, so re-running it while
			// the camera moves makes nodes trade cells and their labels blink in and
			// out. Decide which nodes are labelled only while idle; during a drag or
			// zoom, reuse that set (positions still update every frame) and re-cull
			// once the camera settles. Smallest first so big nodes paint on top.
			let visible: OverlayItem[];
			const frozen = labelledNodes;
			if (!sigma.getGraphState().isIdle && frozen) {
				visible = candidates
					.filter((item) => frozen.has(item.node))
					.sort((a, b) => a.size - b.size);
			} else {
				const gridKept = cullLabelsByGrid(
					candidates,
					LABEL_GRID_CELL_SIZE,
					LABEL_DENSITY,
				);
				const kept = cullOverlappingLabels(gridKept, LABEL_SIZE);
				labelledNodes = new Set(kept.map((item) => item.node));
				visible = kept.reverse();
			}
			for (const item of visible) {
				drawName(ctx, item, true);
			}

			if (selectedDraw) {
				const text = pillText(selectedDraw);
				drawHoverPill(ctx, selectedDraw, text);
				drawName(ctx, selectedDraw, false, text);
			}

			if (hoveredDraw) {
				const text = pillText(hoveredDraw);
				drawHoverPill(ctx, hoveredDraw, text);
				drawName(ctx, hoveredDraw, false, text);
			}
		}

		redrawOverlay = drawOverlay;
		sigma.on("afterRender", drawOverlay);
		drawOverlay();

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
				hoverTimer = null;
				hoveredNode = payload.node;
				hoveredNeighborSet = new Set(g.neighbors(hoveredNode));
				sigmaInst?.scheduleRefresh({ skipIndexation: true });
			}, 150);
		});

		sigma.on("leaveNode", () => {
			if (hoverTimer !== null) {
				clearTimeout(hoverTimer);
				hoverTimer = null;
			}
			container.style.cursor = "default";
			if (hoveredNode !== null) {
				hoveredNode = null;
				hoveredNeighborSet = new Set();
				sigmaInst?.scheduleRefresh({ skipIndexation: true });
			}
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
	if (!hoveredNode) sigmaInst.scheduleRefresh({ skipIndexation: true });
});

// Same reasoning as the selection effect above: dragging the time-window
// slider must only re-run the reducers (which read `windowVisibility` directly)
// and redraw, never rebuild the graph or re-run FA2/Louvain.
$effect(() => {
	void windowVisibility;
	if (!sigmaInst) return;
	sigmaInst.scheduleRefresh({ skipIndexation: true });
});

// Same reasoning again: toggling "Show artists" only needs the reducers
// (which read `showArtists` directly) to re-run and redraw. The graph
// still contains every artist node/edge underneath — only their visibility
// changes — so there's nothing here to rebuild or re-layout.
$effect(() => {
	void showArtists;
	if (!sigmaInst) return;
	sigmaInst.scheduleRefresh({ skipIndexation: true });
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
