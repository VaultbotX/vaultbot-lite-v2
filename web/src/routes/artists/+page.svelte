<script lang="ts">
import { goto } from "$app/navigation";
import { browser } from "$app/environment";
import { untrack } from "svelte";
import ArtistCard from "$lib/ArtistCard.svelte";
import type { PageData } from "./$types";

const CURRENT_KEY = "artists:showCurrent";

let { data }: { data: PageData } = $props();

let q = $state(untrack(() => data.q));
let showCurrent = $state(browser ? localStorage.getItem(CURRENT_KEY) === "true" : false);
let debounce: ReturnType<typeof setTimeout>;

$effect(() => {
	q = data.q;
});

$effect(() => {
	localStorage.setItem(CURRENT_KEY, String(showCurrent));
});

function navigate(newQ: string, newPage: number, newCurrent: boolean) {
	const params = new URLSearchParams({ page: String(newPage) });
	if (newQ.trim()) params.set("q", newQ.trim());
	if (newCurrent) params.set("current", "1");
	goto(`?${params}`, { replaceState: newPage === data.page && newCurrent === data.current });
}

function handleInput(e: Event) {
	const val = (e.target as HTMLInputElement).value;
	q = val;
	clearTimeout(debounce);
	debounce = setTimeout(() => navigate(val, 1, showCurrent), 300);
}

function toggleCurrent() {
	showCurrent = !showCurrent;
	navigate(q, 1, showCurrent);
}

const totalPages = $derived(Math.ceil(data.total / data.pageSize));
</script>

<svelte:head>
	<title>Vaultbot — Artists</title>
</svelte:head>

<div class="page-header">
	<h1>Artists</h1>
	<p class="muted">
		{data.total.toLocaleString()} artist{data.total !== 1 ? "s" : ""}
		{data.q ? `matching "${data.q}"` : ""}
		{data.current ? "in current playlist" : ""}
	</p>
</div>

<div class="controls">
	<input
		type="search"
		placeholder="Search artists…"
		value={q}
		oninput={handleInput}
		class="filter-input mono"
	/>
	<label class="toggle">
		<input
			type="checkbox"
			checked={showCurrent}
			onchange={toggleCurrent}
		/>
		<span>Current playlist only <span class="pill">≤ 2 weeks</span></span>
	</label>
</div>

{#if data.artists.length === 0}
	<p class="empty mono muted">No artists found</p>
{:else}
	<div class="grid">
		{#each data.artists as artist (artist.artist_id)}
			<ArtistCard {artist} />
		{/each}
	</div>

	{#if totalPages > 1}
		<div class="pagination">
			<button
				class="page-btn mono"
				disabled={data.page <= 1}
				onclick={() => navigate(q, data.page - 1, showCurrent)}
			>← Prev</button>
			<span class="page-info mono muted">
				{data.page} / {totalPages}
			</span>
			<button
				class="page-btn mono"
				disabled={data.page >= totalPages}
				onclick={() => navigate(q, data.page + 1, showCurrent)}
			>Next →</button>
		</div>
	{/if}
{/if}

<style>
	.page-header {
		margin-bottom: 1.5rem;
	}

	.page-header h1 {
		font-size: 24px;
		margin-bottom: 0.25rem;
	}

	.page-header p {
		font-size: 13px;
	}

	.controls {
		display: flex;
		align-items: center;
		gap: 1.25rem;
		margin-bottom: 1.5rem;
		flex-wrap: wrap;
	}

	.filter-input {
		flex: 1;
		max-width: 360px;
		min-width: 160px;
		background: var(--surface);
		border: 1px solid var(--border);
		border-radius: var(--radius);
		padding: 0.5rem 0.75rem;
		font-size: 13px;
		color: var(--text);
		outline: none;
		transition: border-color 0.15s;
	}

	.filter-input::placeholder {
		color: var(--text-muted);
	}

	.filter-input:focus {
		border-color: var(--accent);
	}

	.toggle {
		display: flex;
		align-items: center;
		gap: 0.5rem;
		cursor: pointer;
		color: var(--text-muted);
		font-size: 13px;
		user-select: none;
	}

	.toggle input[type="checkbox"] {
		accent-color: var(--accent);
		cursor: pointer;
	}

	.pill {
		display: inline-block;
		font-family: "IBM Plex Mono", monospace;
		font-size: 10px;
		padding: 1px 5px;
		border-radius: 4px;
		background: var(--surface-2);
		border: 1px solid var(--border);
		color: var(--text-muted);
		vertical-align: middle;
	}

	.empty {
		font-size: 13px;
		margin-top: 2rem;
	}

	.grid {
		display: grid;
		grid-template-columns: repeat(auto-fill, minmax(240px, 1fr));
		gap: 1rem;
	}

	.pagination {
		display: flex;
		align-items: center;
		justify-content: center;
		gap: 1rem;
		margin-top: 2rem;
	}

	.page-btn {
		background: var(--surface);
		border: 1px solid var(--border);
		border-radius: var(--radius);
		padding: 0.4rem 0.9rem;
		font-size: 12px;
		color: var(--text-muted);
		cursor: pointer;
		transition: border-color 0.15s, color 0.15s;
	}

	.page-btn:hover:not(:disabled) {
		border-color: var(--accent);
		color: var(--accent);
	}

	.page-btn:disabled {
		opacity: 0.35;
		cursor: default;
	}

	.page-info {
		font-size: 12px;
		min-width: 4rem;
		text-align: center;
	}
</style>
