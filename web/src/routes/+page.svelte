<script lang="ts">
import { goto } from "$app/navigation";
import SpotifyLinkPill from "$lib/SpotifyLinkPill.svelte";
import StatsCharts from "$lib/StatsCharts.svelte";
import type { PageData } from "./$types";

let { data }: { data: PageData } = $props();

const playlistLinks = $derived([
	{ label: "Dynamic", id: data.playlists.dynamic },
	{ label: "Genre Rotation", id: data.playlists.genre },
	{ label: "Highscores", id: data.playlists.highscores },
	{ label: "Throwback", id: data.playlists.throwback },
	{ label: "Variety", id: data.playlists.variety },
]);

function fmt(n: number): string {
	return n.toLocaleString("en-US");
}

function fmtDate(iso: string): string {
	const d = new Date(iso);
	return (
		"Last updated " +
		d.toLocaleDateString("en-US", {
			month: "long",
			day: "numeric",
			year: "numeric",
		}) +
		" at " +
		d.toLocaleTimeString("en-US", {
			hour: "2-digit",
			minute: "2-digit",
			timeZoneName: "short",
		})
	);
}
</script>

<svelte:head>
	<title>Vaultbot :: Stats</title>
</svelte:head>

<p class="meta mono muted">{fmtDate(data.generated_at)}</p>

<div class="summary-grid">
	<div class="stat-card">
		<div class="stat-label">Unique songs</div>
		<div class="stat-value mono">{fmt(data.summary.total_songs)}</div>
	</div>
	<div class="stat-card">
		<div class="stat-label">Archive entries</div>
		<div class="stat-value mono">{fmt(data.summary.total_archive_entries)}</div>
	</div>
	<div class="stat-card">
		<div class="stat-label">Artists</div>
		<div class="stat-value mono">{fmt(data.summary.total_artists)}</div>
	</div>
	<div class="stat-card">
		<div class="stat-label">Genres</div>
		<div class="stat-value mono">{fmt(data.summary.total_genres)}</div>
	</div>
</div>

<div class="playlists">
	<span class="playlists-label mono muted">Playlists</span>
	<div class="playlist-links">
		{#each playlistLinks as playlist (playlist.label)}
			<SpotifyLinkPill type="playlist" id={playlist.id} label={playlist.label} />
		{/each}
	</div>
</div>

<StatsCharts
	{data}
	onGenreClick={(id) => goto(`/graph?node=g:${id}`)}
	onArtistClick={(id) => goto(`/graph?node=a:${id}`)}
/>

<style>
	.meta {
		font-size: 11px;
		margin-bottom: 36px;
	}

	.summary-grid {
		display: grid;
		grid-template-columns: repeat(4, 1fr);
		gap: 16px;
		margin-bottom: 40px;
	}

	.stat-card {
		background: var(--surface);
		border: 1px solid var(--border);
		border-radius: var(--radius);
		padding: 22px 24px;
	}

	.stat-label {
		font-size: 11px;
		font-weight: 500;
		color: var(--text-muted);
		text-transform: uppercase;
		letter-spacing: 0.1em;
		margin-bottom: 10px;
	}

	.stat-value {
		font-size: 30px;
		font-weight: 500;
		letter-spacing: -0.03em;
		color: var(--text);
		line-height: 1;
	}

	.playlists {
		display: flex;
		align-items: center;
		gap: 16px;
		flex-wrap: wrap;
		margin-bottom: 40px;
	}

	.playlists-label {
		font-size: 11px;
		text-transform: uppercase;
		letter-spacing: 0.1em;
	}

	.playlist-links {
		display: flex;
		flex-wrap: wrap;
		gap: 10px;
	}

	@media (max-width: 900px) {
		.summary-grid {
			grid-template-columns: repeat(2, 1fr);
		}
	}

	@media (max-width: 480px) {
		.summary-grid {
			grid-template-columns: 1fr 1fr;
		}

		.stat-value {
			font-size: 24px;
		}
	}
</style>
