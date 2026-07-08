<script lang="ts">
let {
	name,
	count,
	onClick,
}: {
	name: string;
	count?: number;
	onClick: () => void;
} = $props();
</script>

<button type="button" class="chip" onclick={onClick}>
	<span class="chip-name">{name}</span>
	{#if count !== undefined}
		<span class="chip-count mono">{count}</span>
	{/if}
</button>

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
		font-family: inherit;
		color: var(--text);
		cursor: pointer;
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
