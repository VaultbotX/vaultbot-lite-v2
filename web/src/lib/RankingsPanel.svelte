<script lang="ts">
import type { RankingItem, RankingsKind, TimeRange } from "#lib/graph.ts";
import { formatWindowRange, rankingsFetchUrl } from "#lib/graph.ts";

let {
	kind,
	activeWindow,
	allTimeItems = [],
	onSelect,
}: {
	kind: RankingsKind;
	activeWindow: TimeRange | null;
	allTimeItems?: RankingItem[];
	onSelect: (id: number, kind: "genre" | "artist") => void;
} = $props();

let items = $state<RankingItem[]>([]);
let loading = $state(false);
let error = $state(false);
let visibleLimit = $state(50);

$effect(() => {
	visibleLimit = 50;

	if (!activeWindow) {
		items = allTimeItems;
		loading = false;
		error = false;
		return;
	}

	loading = true;
	error = false;

	const timeoutId = setTimeout(() => {
		fetch(rankingsFetchUrl(kind, activeWindow))
			.then((r) => {
				if (!r.ok) throw new Error(`Failed to load: ${r.status}`);
				return r.json() as Promise<RankingItem[]>;
			})
			.then((data) => {
				items = data;
				loading = false;
			})
			.catch(() => {
				error = true;
				loading = false;
			});
	}, 250);

	return () => clearTimeout(timeoutId);
});

const visibleItems = $derived(items.slice(0, visibleLimit));
const hasMore = $derived(visibleLimit < items.length);

function loadMore(): void {
	if (hasMore) {
		visibleLimit += 50;
	}
}

function sentinel(node: HTMLElement) {
	const observer = new IntersectionObserver(
		(entries) => {
			if (entries[0]?.isIntersecting) {
				loadMore();
			}
		},
		{ rootMargin: "200px" },
	);
	observer.observe(node);
	return {
		destroy() {
			observer.disconnect();
		},
	};
}
</script>

<h2 class="title">{kind === "genres" ? "Top Genres" : "Top Artists"}</h2>

{#if activeWindow}
	<p class="window-note mono muted">
		Showing {formatWindowRange(activeWindow[0], activeWindow[1])}
	</p>
{/if}

{#if loading}
	<div class="status muted mono">Loading rankings…</div>
{:else if error}
	<div class="status muted mono">Failed to load rankings.</div>
{:else if items.length === 0}
	<p class="empty mono muted">
		{activeWindow
			? `No ${kind} found in this time period`
			: `No ${kind} found`}
	</p>
{:else}
	<section class="block">
		<div class="summary-line mono muted">
			{items.length.toLocaleString()} total {kind}
		</div>
		<table>
			<thead>
				<tr>
					<th class="rank-col mono">#</th>
					<th>{kind === "genres" ? "Genre" : "Artist"}</th>
					<th class="right mono">Entries</th>
				</tr>
			</thead>
			<tbody>
				{#each visibleItems as item (item.id)}
					<tr>
						<td class="rank-col mono muted">#{item.rank}</td>
						<td>
							<button
								type="button"
								class="inner-link"
								onclick={() =>
									onSelect(
										item.id,
										kind === "genres" ? "genre" : "artist",
									)}
							>
								{item.name}
							</button>
						</td>
						<td class="right mono">{item.entries.toLocaleString()}</td>
					</tr>
				{/each}
			</tbody>
		</table>
		{#if hasMore}
			<div use:sentinel class="sentinel">
				<span class="mono muted loading-more">Loading more…</span>
			</div>
		{/if}
	</section>
{/if}

<style>
	.title {
		font-size: 20px;
		margin-bottom: 0.5rem;
	}

	.window-note {
		font-size: 11px;
		margin: 0 0 1.25rem;
	}

	.status {
		font-size: 13px;
		padding: 2rem 0;
		text-align: center;
	}

	.summary-line {
		font-size: 11px;
		margin-bottom: 0.75rem;
	}

	.block {
		margin-bottom: 1.5rem;
	}

	.empty {
		font-size: 12px;
		padding: 1rem 0;
	}

	table {
		width: 100%;
		border-collapse: collapse;
		font-size: 13px;
	}

	thead th {
		text-align: left;
		font-size: 11px;
		font-weight: 500;
		text-transform: uppercase;
		letter-spacing: 0.05em;
		color: var(--text-muted);
		padding: 0 0 0.5rem;
		border-bottom: 1px solid var(--border);
	}

	thead th.right {
		text-align: right;
	}

	thead th.rank-col,
	tbody td.rank-col {
		width: 48px;
		text-align: left;
	}

	tbody tr {
		border-bottom: 1px solid var(--border);
	}

	tbody tr:last-child {
		border-bottom: none;
	}

	tbody td {
		padding: 0.5rem 0;
		vertical-align: middle;
	}

	tbody td.right {
		text-align: right;
		white-space: nowrap;
	}

	.inner-link {
		display: inline;
		background: none;
		border: none;
		padding: 0;
		font: inherit;
		text-align: left;
		cursor: pointer;
		color: var(--text);
		transition: color 0.15s;
	}

	.inner-link:hover {
		color: var(--accent);
		text-decoration: none;
	}

	.sentinel {
		padding: 1rem 0;
		text-align: center;
	}

	.loading-more {
		font-size: 11px;
	}
</style>
