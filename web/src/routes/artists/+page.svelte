<script lang="ts">
import type { PageData } from "./$types";

let { data }: { data: PageData } = $props();

let filter = $state("");

const filtered = $derived(
	filter.trim() === ""
		? data.artists
		: data.artists.filter((a) =>
				a.name.toLowerCase().includes(filter.trim().toLowerCase()),
			),
);
</script>

<svelte:head>
	<title>Vaultbot — Artists</title>
</svelte:head>

<div class="page-header">
	<h1>Artists</h1>
	<p class="muted">{data.artists.length.toLocaleString()} artists in the archive</p>
</div>

<div class="controls">
	<input
		type="search"
		placeholder="Filter artists…"
		bind:value={filter}
		class="filter-input mono"
	/>
	{#if filter.trim() !== ""}
		<span class="result-count mono muted">{filtered.length.toLocaleString()} result{filtered.length !== 1 ? "s" : ""}</span>
	{/if}
</div>

{#if filtered.length === 0}
	<p class="empty mono muted">No artists match "{filter}"</p>
{:else}
	<div class="grid">
		{#each filtered as artist (artist.artist_id)}
			<a href="/artists/{artist.artist_id}" class="artist-card">
				<div class="artist-name">{artist.name}</div>
				<div class="artist-stats">
					<span class="stat">
						<span class="stat-value mono">{artist.unique_songs.toLocaleString()}</span>
						<span class="stat-label muted">songs</span>
					</span>
					<span class="stat">
						<span class="stat-value mono">{artist.archive_count.toLocaleString()}</span>
						<span class="stat-label muted">entries</span>
					</span>
					<span class="stat">
						<span class="stat-value mono">{artist.genre_count.toLocaleString()}</span>
						<span class="stat-label muted">genres</span>
					</span>
				</div>
			</a>
		{/each}
	</div>
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
		gap: 1rem;
		margin-bottom: 1.5rem;
	}

	.filter-input {
		flex: 1;
		max-width: 400px;
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

	.result-count {
		font-size: 12px;
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

	.artist-card {
		background: var(--surface);
		border: 1px solid var(--border);
		border-radius: var(--radius);
		padding: 1rem 1.25rem;
		text-decoration: none;
		color: var(--text);
		transition: border-color 0.15s, background 0.15s;
		display: flex;
		flex-direction: column;
		gap: 0.75rem;
	}

	.artist-card:hover {
		border-color: var(--accent);
		background: var(--surface-2);
		text-decoration: none;
	}

	.artist-name {
		font-size: 14px;
		font-weight: 500;
		line-height: 1.3;
		word-break: break-word;
	}

	.artist-stats {
		display: flex;
		gap: 1rem;
	}

	.stat {
		display: flex;
		flex-direction: column;
		gap: 2px;
	}

	.stat-value {
		font-size: 15px;
		font-weight: 500;
		color: var(--text);
		line-height: 1;
	}

	.stat-label {
		font-size: 10px;
		text-transform: uppercase;
		letter-spacing: 0.06em;
	}
</style>
