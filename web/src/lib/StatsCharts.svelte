<script lang="ts">
import { onMount } from "svelte";
import { fmtMonth, treemapColor } from "$lib/stats";
import type { StatsData } from "../routes/api/stats/+server";

let { data }: { data: StatsData } = $props();

let timeEl: HTMLCanvasElement | undefined;
let artistsEl: HTMLCanvasElement | undefined;
let genresEl: HTMLCanvasElement | undefined;

const ACCENT = "#7c6af7";
const ACCENT_FILL = "rgba(124, 106, 247, 0.12)";
const ACCENT_HOVER = "#9d90f9";
const TOOLTIP_OPTS = {
	backgroundColor: "#1a1a22",
	borderColor: "#1f1f2c",
	borderWidth: 1,
	titleColor: "#e2e2f0",
	bodyColor: "#9090c0",
	padding: 12,
};

function fmt(n: number): string {
	return n.toLocaleString("en-US");
}

onMount(() => {
	Promise.all([import("chart.js"), import("chartjs-chart-treemap")]).then(
		([chartjs, treemap]) => {
			const {
				Chart,
				CategoryScale,
				LinearScale,
				PointElement,
				LineElement,
				LineController,
				BarElement,
				BarController,
				Filler,
				Tooltip,
			} = chartjs;
			const { TreemapController, TreemapElement } = treemap;

			Chart.register(
				CategoryScale,
				LinearScale,
				PointElement,
				LineElement,
				LineController,
				BarElement,
				BarController,
				Filler,
				Tooltip,
				TreemapController,
				TreemapElement,
			);

			Chart.defaults.color = "#6060a0";
			Chart.defaults.font.family = "'IBM Plex Sans', sans-serif";
			Chart.defaults.font.size = 12;
			Chart.defaults.borderColor = "#1f1f2c";

			if (!timeEl || !artistsEl || !genresEl) return;

			const MONO = { family: "'IBM Plex Mono', monospace", size: 11 };

			const timeChart = new Chart(timeEl, {
				type: "line",
				data: {
					labels: data.songs_over_time.map((d) => fmtMonth(d.month)),
					datasets: [
						{
							label: "Archive entries",
							data: data.songs_over_time.map((d) => d.count),
							fill: true,
							borderColor: ACCENT,
							backgroundColor: ACCENT_FILL,
							borderWidth: 2,
							tension: 0.35,
							pointRadius: 2,
							pointHoverRadius: 5,
							pointBackgroundColor: ACCENT,
							pointHoverBackgroundColor: ACCENT_HOVER,
						},
					],
				},
				options: {
					responsive: true,
					maintainAspectRatio: false,
					interaction: { mode: "index", intersect: false },
					plugins: {
						legend: { display: false },
						tooltip: TOOLTIP_OPTS,
					},
					scales: {
						x: {
							grid: { color: "#1a1a24" },
							ticks: { maxTicksLimit: 12, maxRotation: 0, font: MONO },
						},
						y: {
							grid: { color: "#1a1a24" },
							ticks: {
								font: MONO,
								callback: (v) => fmt(v as number),
							},
							beginAtZero: true,
						},
					},
				},
			});

			const artistMax = data.top_artists[0]?.song_count ?? 1;
			const artists = [...data.top_artists].reverse();

			const artistsChart = new Chart(artistsEl, {
				type: "bar",
				data: {
					labels: artists.map((a) => a.name),
					datasets: [
						{
							data: artists.map((a) => a.song_count),
							backgroundColor: artists.map((a) =>
								treemapColor(a.song_count / artistMax),
							),
							borderRadius: 3,
							borderSkipped: false,
						},
					],
				},
				options: {
					indexAxis: "y",
					responsive: true,
					maintainAspectRatio: false,
					plugins: {
						legend: { display: false },
						tooltip: {
							...TOOLTIP_OPTS,
							callbacks: {
								label: (ctx) =>
									` ${fmt(ctx.parsed.x)} song${ctx.parsed.x !== 1 ? "s" : ""}`,
							},
						},
					},
					scales: {
						x: {
							grid: { color: "#1a1a24" },
							ticks: { font: MONO, callback: (v) => fmt(v as number) },
							beginAtZero: true,
						},
						y: {
							grid: { display: false },
							ticks: { font: { size: 12 }, color: "#b0b0d0" },
						},
					},
				},
			});

			const genreMax = data.genre_distribution[0]?.song_count ?? 1;

			// chartjs-chart-treemap attaches the source object as ctx.raw._data at runtime
			// but the TypeScript types don't expose it — cast through unknown where needed.
			type TreemapCtx = {
				type: string;
				raw: { _data: { name: string; song_count: number } };
			};

			const genresChart = new Chart(genresEl, {
				type: "treemap",
				data: {
					datasets: [
						{
							// @ts-expect-error — chartjs-chart-treemap dataset shape
							tree: data.genre_distribution,
							key: "song_count",
							labels: {
								display: true,
								align: "center",
								position: "middle",
								color: "rgba(226, 226, 240, 0.90)",
								font: [
									{
										family: "'IBM Plex Mono', monospace",
										size: 11,
										weight: 500,
									},
									{ family: "'IBM Plex Mono', monospace", size: 10 },
								],
								formatter: (ctx: unknown) => {
									const d = (ctx as TreemapCtx).raw._data;
									return [d.name, fmt(d.song_count)];
								},
							},
							backgroundColor: (ctx: unknown) => {
								const c = ctx as TreemapCtx;
								if (c.type !== "data") return "transparent";
								const ratio = (c.raw._data?.song_count ?? 0) / genreMax;
								return treemapColor(ratio);
							},
							borderColor: "#0c0c10",
							borderWidth: 2,
							spacing: 1,
						},
					],
				},
				options: {
					responsive: true,
					maintainAspectRatio: false,
					plugins: {
						legend: { display: false },
						tooltip: {
							...TOOLTIP_OPTS,
							callbacks: {
								title: (items: unknown[]) => {
									const d = (items[0] as TreemapCtx).raw._data;
									return d.name;
								},
								label: (item: unknown) => {
									const d = (item as TreemapCtx).raw._data;
									return ` ${fmt(d.song_count)} song${d.song_count !== 1 ? "s" : ""}`;
								},
							},
						},
					},
				},
			});

			return () => {
				timeChart.destroy();
				artistsChart.destroy();
				genresChart.destroy();
			};
		},
	);
});
</script>

<div class="charts">
	<div class="chart-card">
		<div class="chart-header">
			<span class="chart-title">Archive entries over time</span>
		</div>
		<div class="chart-wrap chart-wrap--tall">
			<canvas bind:this={timeEl}></canvas>
		</div>
	</div>

	<div class="chart-row">
		<div class="chart-card">
			<div class="chart-header">
				<span class="chart-title">Top artists</span>
			</div>
			<div class="chart-wrap chart-wrap--medium">
				<canvas bind:this={artistsEl}></canvas>
			</div>
		</div>

		<div class="chart-card">
			<div class="chart-header">
				<span class="chart-title">Genre distribution</span>
			</div>
			<div class="chart-wrap chart-wrap--treemap">
				<canvas bind:this={genresEl}></canvas>
			</div>
		</div>
	</div>
</div>

<style>
	.charts {
		display: flex;
		flex-direction: column;
		gap: 24px;
	}

	.chart-card {
		background: var(--surface);
		border: 1px solid var(--border);
		border-radius: var(--radius);
		padding: 28px;
	}

	.chart-header {
		margin-bottom: 24px;
	}

	.chart-title {
		font-size: 13px;
		font-weight: 600;
		text-transform: uppercase;
		letter-spacing: 0.08em;
		color: var(--text-muted);
	}

	.chart-row {
		display: grid;
		grid-template-columns: 1fr 1fr;
		gap: 24px;
	}

	.chart-wrap {
		position: relative;
	}

	.chart-wrap--tall {
		height: 300px;
	}

	.chart-wrap--medium {
		height: 380px;
	}

	.chart-wrap--treemap {
		height: 420px;
	}

	@media (max-width: 900px) {
		.chart-row {
			grid-template-columns: 1fr;
		}
	}
</style>
