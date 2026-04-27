<script lang="ts">
import type { ArtistSummary } from "../routes/api/artists/+server";

let { artist }: { artist: ArtistSummary } = $props();
</script>

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

<style>
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
