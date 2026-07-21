<script lang="ts">
import GenreChip from "$lib/GenreChip.svelte";
import type { TimeRange } from "$lib/graph";
import { formatRank, formatWindowRange } from "$lib/graph";
import { spotifyUrl } from "$lib/spotify";
import type { GenreDetail } from "../routes/api/genres/[id]/+server";

let {
	data,
	activeWindow,
	onSelectGenre,
	onSelectArtist,
}: {
	data: GenreDetail;
	activeWindow: TimeRange | null;
	onSelectGenre: (id: number) => void;
	onSelectArtist: (id: number) => void;
} = $props();
</script>

<h2 class="title">{data.genre_name}</h2>
<p class="rank mono muted">{formatRank("genre", data.rank, data.rank_total)}</p>
{#if activeWindow}
	<p class="window-note mono muted">
		Showing {formatWindowRange(activeWindow[0], activeWindow[1])}
	</p>
{/if}

<section class="block">
	<h3>Artists</h3>
	{#if data.artists.length === 0}
		<p class="empty mono muted">
			{activeWindow ? "No artists in this time period" : "No artists found"}
		</p>
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
							<button type="button" class="inner-link" onclick={() => onSelectArtist(artist.artist_id)}
								>{artist.name}</button
							>
						</td>
						<td class="right mono">{artist.archive_count.toLocaleString()}</td>
					</tr>
				{/each}
			</tbody>
		</table>
	{/if}
</section>

<section class="block">
	<h3>Top Tracks</h3>
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
								<button
									type="button"
									class="inner-link muted-link"
									onclick={() => onSelectArtist(track.artist_ids[i])}>{name}</button
								>
							{/each}
						</td>
						<td class="right mono">{track.occurrences.toLocaleString()}</td>
					</tr>
				{/each}
			</tbody>
		</table>
	{/if}
</section>

{#if data.connected_genres.length > 0}
	<section class="block">
		<h3>Related Genres</h3>
		<div class="chips">
			{#each data.connected_genres as genre}
				<GenreChip
					name={genre.name}
					count={genre.shared_archive_count}
					onClick={() => onSelectGenre(genre.genre_id)}
				/>
			{/each}
		</div>
	</section>
{/if}

<style>
	.title {
		font-size: 20px;
		margin-bottom: 1.25rem;
	}

	.rank {
		font-size: 11px;
		margin: -1rem 0 1.25rem;
	}

	.window-note {
		font-size: 11px;
		margin: -1rem 0 1.25rem;
	}

	.block {
		margin-bottom: 1.5rem;
	}

	.block h3 {
		font-size: 13px;
		font-weight: 600;
		text-transform: uppercase;
		letter-spacing: 0.06em;
		color: var(--text-muted);
		margin-bottom: 0.75rem;
		padding-bottom: 0.5rem;
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

	.muted-link {
		color: var(--text-muted);
	}

	.track-name {
		padding-right: 0.5rem;
		overflow-wrap: break-word;
	}

	.artist-list {
		font-size: 12px;
		padding-right: 0.5rem;
	}

	.chips {
		display: flex;
		flex-wrap: wrap;
		gap: 0.5rem;
	}
</style>
