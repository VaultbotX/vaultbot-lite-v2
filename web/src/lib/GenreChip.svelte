<script lang="ts">
let {
	genre_id,
	name,
	count,
}: {
	genre_id: number;
	name: string;
	count?: number;
} = $props();
</script>

<a href="/genres/{genre_id}" class="chip">
	<span class="chip-name">{name}</span>
	{#if count !== undefined}
		<span class="chip-count mono">{count}</span>
	{/if}
</a>

<style>
	.chip {
		display: inline-flex;
		align-items: center;
		gap: 0.4rem;
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

	.chip-count {
		font-size: 11px;
		color: var(--text-muted);
		transition: color 0.15s;
	}

	.chip:hover .chip-count {
		color: var(--accent);
	}
</style>
