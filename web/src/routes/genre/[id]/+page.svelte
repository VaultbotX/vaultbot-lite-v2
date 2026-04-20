<script lang="ts">
	import type { PageData } from './$types';

	let { data }: { data: PageData } = $props();
</script>

<svelte:head>
	<title>Vaultbot — Genre</title>
</svelte:head>

<div class="page-header">
	<a href="/" class="back mono">← Back to graph</a>
	<h1>Genre</h1>
	<p class="muted">Top artists and tracks for this genre in the archive.</p>
</div>

<div class="grid">
	<section class="card">
		<h2>Artists</h2>
		<p class="muted placeholder mono">Artist data coming soon</p>
	</section>

	<section class="card">
		<h2>Top Tracks</h2>
		<p class="muted placeholder mono">Track data coming soon</p>
	</section>
</div>

<style>
	.page-header {
		margin-bottom: 1.5rem;
	}

	.back {
		display: inline-block;
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
	}

	.card h2 {
		font-size: 14px;
		font-weight: 600;
		margin-bottom: 1rem;
		padding-bottom: 0.75rem;
		border-bottom: 1px solid var(--border);
	}

	.placeholder {
		font-size: 12px;
	}

	@media (max-width: 700px) {
		.grid {
			grid-template-columns: 1fr;
		}
	}
</style>
