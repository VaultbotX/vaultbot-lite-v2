<script lang="ts">
import type { PageData } from "./$types";

let { data }: { data: PageData } = $props();

function spotifyArtistUrl(spotifyId: string): string {
	return `spotify:artist:${spotifyId}`;
}

function spotifyTrackUrl(spotifyId: string): string {
	return `spotify:track:${spotifyId}`;
}
</script>

<svelte:head>
	<title>Vaultbot — {data.artist_name || "Artist"}</title>
</svelte:head>

<div class="page-header">
	<button onclick={() => history.back()} class="back mono">← Back</button>
	{#if data.notFound}
		<h1>Artist not found</h1>
	{:else}
		<div class="title-row">
			<h1>{data.artist_name}</h1>
			<a
				href={spotifyArtistUrl(data.spotify_id)}
				target="_blank"
				rel="noopener noreferrer"
				class="spotify-btn mono"
			>Open in Spotify ↗</a>
		</div>
	{/if}
</div>

{#if !data.notFound}
	{#if data.genres.length > 0}
		<section class="card genres-section">
			<h2>Genres</h2>
			<div class="chips">
				{#each data.genres as genre}
					<a href="/genres/{genre.genre_id}" class="chip">{genre.name}</a>
				{/each}
			</div>
		</section>
	{/if}

	<section class="card songs-section">
		<h2>Songs <span class="count mono muted">({data.songs.length.toLocaleString()})</span></h2>
		{#if data.songs.length === 0}
			<p class="empty mono muted">No songs found</p>
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
					{#each data.songs as song}
						<tr>
							<td class="track-name">
								<a
									href={spotifyTrackUrl(song.spotify_id)}
									target="_blank"
									rel="noopener noreferrer"
									class="track-link"
								>{song.name}</a>
							</td>
							<td class="artist-list muted">
								{#each song.artist_names as name, i}
									{#if i > 0}<span>, </span>{/if}
									<a href="/artists/{song.artist_ids[i]}" class="artist-link">{name}</a>
								{/each}
							</td>
							<td class="right mono">{song.occurrences.toLocaleString()}</td>
						</tr>
					{/each}
				</tbody>
			</table>
		{/if}
	</section>
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

	.title-row {
		display: flex;
		align-items: baseline;
		gap: 1rem;
		flex-wrap: wrap;
	}

	.page-header h1 {
		font-size: 24px;
		margin-bottom: 0.25rem;
	}

	.spotify-btn {
		font-size: 11px;
		color: var(--text-muted);
		border: 1px solid var(--border);
		border-radius: 999px;
		padding: 0.2rem 0.6rem;
		text-decoration: none;
		transition: color 0.15s, border-color 0.15s;
		white-space: nowrap;
	}

	.spotify-btn:hover {
		color: var(--accent);
		border-color: var(--accent);
		text-decoration: none;
	}

	.genres-section {
		margin-bottom: 1.5rem;
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

	.count {
		text-transform: none;
		letter-spacing: 0;
		font-weight: 400;
		font-size: 12px;
	}

	.chips {
		display: flex;
		flex-wrap: wrap;
		gap: 0.5rem;
	}

	.chip {
		display: inline-flex;
		align-items: center;
		padding: 0.3rem 0.65rem;
		background: var(--surface-2);
		border: 1px solid var(--border);
		border-radius: 999px;
		font-size: 12px;
		color: var(--text);
		transition: border-color 0.15s, color 0.15s;
		text-decoration: none;
	}

	.chip:hover {
		border-color: var(--accent);
		color: var(--accent);
		text-decoration: none;
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

	.track-link {
		color: var(--text);
		transition: color 0.15s;
	}

	.track-link:hover {
		color: var(--accent);
		text-decoration: none;
	}

	.artist-link {
		color: var(--text-muted);
		transition: color 0.15s;
	}

	.artist-link:hover {
		color: var(--accent);
		text-decoration: none;
	}

	.track-name {
		padding-right: 0.5rem;
		white-space: nowrap;
		overflow: hidden;
		text-overflow: ellipsis;
		max-width: 260px;
	}

	.artist-list {
		font-size: 12px;
		padding-right: 0.5rem;
	}
</style>
