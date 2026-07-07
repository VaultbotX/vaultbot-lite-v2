<script lang="ts">
import { spotifyUrl } from "$lib/spotify";

let {
	type,
	id,
	label,
}: {
	type: "artist" | "track" | "playlist";
	id: string;
	label: string;
} = $props();
</script>

<a
	class="spotify-pill mono"
	href={spotifyUrl(type, id)}
	target="_blank"
	rel="noopener noreferrer"
>
	{label}
	<span class="arrow" aria-hidden="true">↗</span>
</a>

<style>
	.spotify-pill {
		display: inline-flex;
		align-items: center;
		gap: 0.4rem;
		padding: 6px 14px;
		font-size: 13px;
		color: var(--text);
		background: var(--surface);
		border: 1px solid var(--border);
		border-radius: 999px;
		transition:
			border-color 0.15s,
			color 0.15s;
		white-space: nowrap;
	}

	.spotify-pill:hover {
		color: var(--accent);
		border-color: var(--accent);
		text-decoration: none;
	}

	.arrow {
		font-size: 11px;
	}
</style>
