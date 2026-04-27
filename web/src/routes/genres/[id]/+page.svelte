<script lang="ts">
import GenreChip from "$lib/GenreChip.svelte";
import { spotifyUrl } from "$lib/spotify";
import type { PageData } from "./$types";

let { data }: { data: PageData } = $props();
</script>

<svelte:head>
	<title>Vaultbot — {data.genre_name || "Genre"}</title>
</svelte:head>

<div class="page-header">
	<button onclick={() => history.back()} class="back mono">← Back</button>
	{#if data.notFound}
		<h1>Genre not found</h1>
	{:else}
		<h1>{data.genre_name}</h1>
	{/if}
</div>

{#if !data.notFound}
	<div class="grid">
		<section class="card">
			<h2>Artists</h2>
			{#if data.artists.length === 0}
				<p class="empty mono muted">No artists found</p>
			{:else}
				<table>
					<thead>
						<tr>
							<th>Artist</th>
							<th class="right mono">Entries</th>
						</tr>
					</thead>
					<tbody>
						{#each data.artists as artist}
							<tr>
								<td>
									<a href="/artists/{artist.artist_id}" class="inner-link">{artist.name}</a>
								</td>
								<td class="right mono">{artist.archive_count.toLocaleString()}</td>
							</tr>
						{/each}
					</tbody>
				</table>
			{/if}
		</section>

		<section class="card">
			<h2>Top Tracks</h2>
			{#if data.tracks.length === 0}
				<p class="empty mono muted">No tracks found</p>
			{:else}
				<table>
					<thead>
						<tr>
							<th>Track</th>
							<th>Artists</th>
							<th class="right mono">Entries</th>
						</tr>
					</thead>
					<tbody>
						{#each data.tracks as track}
							<tr>
								<td class="track-name">
									<a
										href={spotifyUrl("track", track.spotify_id)}
										target="_blank"
										rel="noopener noreferrer"
										class="inner-link"
									>{track.name}</a>
								</td>
								<td class="artist-list muted">
									{#each track.artist_names as name, i}
										{#if i > 0}<span>, </span>{/if}
										<a href="/artists/{track.artist_ids[i]}" class="inner-link muted-link">{name}</a>
									{/each}
								</td>
								<td class="right mono">{track.occurrences.toLocaleString()}</td>
							</tr>
						{/each}
					</tbody>
				</table>
			{/if}
		</section>
	</div>

	{#if data.connected_genres.length > 0}
		<section class="card related">
			<h2>Related Genres</h2>
			<div class="chips">
				{#each data.connected_genres as genre}
					<GenreChip genre_id={genre.genre_id} name={genre.name} count={genre.shared_artist_count} />
				{/each}
			</div>
		</section>
	{/if}
{/if}

<style>
	.page-header {
		margin-bottom: 1.5rem;
	}

	.back {
		display: inline-block;
		background: none;
		border: none;
		padding: 0;
		cursor: pointer;
		font-size: 12px;
		color: var(--text-muted);
		margin-bottom: 0.75rem;
		transition: color 0.15s;
	}

	.back:hover {
		color: var(--accent);
		text-decoration: none;
	}

	.page-header h1 {
		font-size: 24px;
		margin-bottom: 0.25rem;
	}

	.grid {
		display: grid;
		grid-template-columns: 1fr 1fr;
		gap: 1.5rem;
		align-items: start;
	}

	.card h2 {
		font-size: 13px;
		font-weight: 600;
		text-transform: uppercase;
		letter-spacing: 0.06em;
		color: var(--text-muted);
		margin-bottom: 1rem;
		padding-bottom: 0.75rem;
		border-bottom: 1px solid var(--border);
	}

	.empty {
		font-size: 12px;
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

	tbody tr {
		border-bottom: 1px solid var(--border);
	}

	tbody tr:last-child {
		border-bottom: none;
	}

	tbody td {
		padding: 0.5rem 0;
		vertical-align: top;
	}

	tbody td.right {
		text-align: right;
		white-space: nowrap;
	}

	.inner-link {
		color: var(--text);
		transition: color 0.15s;
	}

	.inner-link:hover {
		color: var(--accent);
		text-decoration: none;
	}

	.muted-link {
		color: var(--text-muted);
	}

	.track-name {
		padding-right: 0.5rem;
		white-space: nowrap;
		overflow: hidden;
		text-overflow: ellipsis;
		max-width: 180px;
	}

	.artist-list {
		font-size: 12px;
		padding-right: 0.5rem;
	}

	.related {
		margin-top: 1.5rem;
	}

	.chips {
		display: flex;
		flex-wrap: wrap;
		gap: 0.5rem;
	}

	@media (max-width: 700px) {
		.grid {
			grid-template-columns: 1fr;
		}
	}
</style>
