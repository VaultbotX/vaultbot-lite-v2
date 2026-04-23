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
	refresh(): void;
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
	commList.forEach((commId, ci) => {
		const r = centerR * Math.sqrt(ci / commList.length);
		const cx = r * Math.cos(ci * GOLDEN_ANGLE);
		const cy = r * Math.sin(ci * GOLDEN_ANGLE);
		const members = groups.get(commId) ?? [];
		members.forEach((node, mi) => {
			const a = (2 * Math.PI * mi) / members.length;
			g.setNodeAttribute(node, "x", cx + memberR * Math.cos(a));
			g.setNodeAttribute(node, "y", cy + memberR * Math.sin(a));
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
	if (!Sigma || !fa2 || !edgeCurve || !container || !nodeBorder || !sigmaRendering) return;

	loading = true;

	const { drawDiscNodeLabel } = sigmaRendering;

	// Border program: 3px black outer ring, community color fill.
	// Border color is read from the `borderColor` attribute so the nodeReducer
	// can dim it alongside the fill when a neighbor is not highlighted.
	const nodeBorderProgram = nodeBorder.createNodeBorderProgram({
		borders: [
			{
				size: { value: 3, mode: "pixels" },
				color: { attribute: "borderColor", defaultValue: "#000000" },
			},
			{ size: { fill: true }, color: { attribute: "color", defaultValue: "#7c6af7" } },
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
			const xDeltaCoord = Math.sqrt(Math.abs(radius ** 2 - (boxHeight / 2) ** 2));

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
				barnesHutOptimize: false,
			},
			getEdgeWeight: "shared",
		});

		// Hover state — closed over by the reducers and event handlers below.
		let hoveredNode: string | null = null;
		let neighborSet = new Set<string>();
		let hoverTimer: ReturnType<typeof setTimeout> | null = null;

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
			defaultDrawNodeHover: drawDarkNodeHover,
			edgeProgramClasses: { curve: edgeCurve },
			nodeProgramClasses: { circle: nodeBorderProgram },
			// Dim nodes/edges that are not part of the hovered node's neighborhood.
			nodeReducer: (node: unknown, data: unknown) => {
				const d = data as Record<string, unknown>;
				if (!hoveredNode || node === hoveredNode || neighborSet.has(node as string)) {
					return d;
				}
				return { ...d, color: "#1e1e28", borderColor: "#1e1e28", label: "" };
			},
			edgeReducer: (edge: unknown, data: unknown) => {
				const d = data as Record<string, unknown>;
				if (!hoveredNode || g.hasExtremity(edge as string, hoveredNode)) {
					return d;
				}
				return { ...d, hidden: true };
			},
		});

		sigmaInst.on("clickNode", (payload) => {
			const node = payload.node as string;
			onNodeTap(g.getNodeAttribute(node, "genreId") as number);
		});

		sigmaInst.on("enterNode", (payload) => {
			container.style.cursor = "pointer";
			if (hoverTimer !== null) clearTimeout(hoverTimer);
			hoverTimer = setTimeout(() => {
				hoveredNode = payload.node as string;
				neighborSet = new Set(g.neighbors(hoveredNode));
				sigmaInst?.refresh();
			}, 150);
		});

		sigmaInst.on("leaveNode", () => {
			if (hoverTimer !== null) {
				clearTimeout(hoverTimer);
				hoverTimer = null;
			}
			hoveredNode = null;
			neighborSet = new Set();
			sigmaInst?.refresh();
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
		background: var(--bg);
	}
</style>
