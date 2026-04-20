<script lang="ts">
import type { PageData } from "./$types";

let { data }: { data: PageData } = $props();
</script>

<svelte:head>
	<title>Vaultbot — Genre Graph</title>
</svelte:head>

<div class="page-header">
	<h1>Genre Graph</h1>
	<p class="muted">Explore connections between genres in the archive. Click a node to drill down.</p>
</div>

<div class="toolbar card">
	<label>
		<input type="checkbox" />
		<span>Show genres with fewer than 3 artists</span>
	</label>
</div>

<div class="graph-container card">
	<div class="graph-placeholder mono muted">Graph visualization coming soon</div>
</div>

<style>
	.page-header {
		margin-bottom: 1.5rem;
	}

	.page-header h1 {
		font-size: 24px;
		margin-bottom: 0.25rem;
	}

	.toolbar {
		margin-bottom: 1rem;
		display: flex;
		align-items: center;
		gap: 1rem;
		padding: 0.75rem 1rem;
	}

	.toolbar label {
		display: flex;
		align-items: center;
		gap: 0.5rem;
		cursor: pointer;
		color: var(--text-muted);
		font-size: 13px;
	}

	.toolbar input[type='checkbox'] {
		accent-color: var(--accent);
	}

	.graph-container {
		height: 75vh;
		min-height: 500px;
		display: flex;
		align-items: center;
		justify-content: center;
	}

	.graph-placeholder {
		font-size: 13px;
	}
</style>
