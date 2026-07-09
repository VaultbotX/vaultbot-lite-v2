<script lang="ts">
import GenreChip from "$lib/GenreChip.svelte";
import SpotifyLinkPill from "$lib/SpotifyLinkPill.svelte";
import { spotifyUrl } from "$lib/spotify";
import type { ArtistDetail } from "../routes/api/artists/[id]/+server";

let {
	data,
	onSelectGenre,
	onSelectArtist,
}: {
	data: ArtistDetail;
	onSelectGenre: (id: number) => void;
	onSelectArtist: (id: number) => void;
} = $props();
</script>

<div class="title-row">
	<h2 class="title">{data.artist_name}</h2>
	<SpotifyLinkPill type="artist" id={data.spotify_id} label="Open in Spotify" />
</div>

{#if data.genres.length > 0}
	<section class="block">
		<h3>Genres</h3>
		<div class="chips">
			{#each data.genres as genre}
				<GenreChip name={genre.name} onClick={() => onSelectGenre(genre.genre_id)} />
			{/each}
		</div>
	</section>
{/if}

<section class="block">
	<h3>Songs <span class="count mono muted">({data.songs.length.toLocaleString()})</span></h3>
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
								href={spotifyUrl("track", song.spotify_id)}
								target="_blank"
								rel="noopener noreferrer"
								class="inner-link"
							>{song.name}</a>
						</td>
						<td class="artist-list muted">
							{#each song.artist_names as name, i}
								{#if i > 0}<span>, </span>{/if}
								<button
									type="button"
									class="inner-link muted-link"
									onclick={() => onSelectArtist(song.artist_ids[i])}>{name}</button
								>
							{/each}
						</td>
						<td class="right mono">{song.occurrences.toLocaleString()}</td>
					</tr>
				{/each}
			</tbody>
		</table>
	{/if}
</section>

{#if data.connected_artists.length > 0}
	<section class="block">
		<h3>Collaborators</h3>
		<div class="chips">
			{#each data.connected_artists as artist}
				<GenreChip
					name={artist.name}
					count={artist.shared_song_count}
					onClick={() => onSelectArtist(artist.artist_id)}
				/>
			{/each}
		</div>
	</section>
{/if}

<style>
	.title-row {
		display: flex;
		align-items: baseline;
		gap: 1rem;
		flex-wrap: wrap;
		margin-bottom: 1.25rem;
	}

	.title {
		font-size: 20px;
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
</style>
