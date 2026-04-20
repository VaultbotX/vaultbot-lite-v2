<script lang="ts">
	import '../app.css';

	let { children } = $props();
</script>

<svelte:head>
	<title>Vaultbot</title>
</svelte:head>

<header>
	<div class="header-inner">
		<div class="brand">
			<img src="/vaultbot.png" alt="Vaultbot logo" width="44" height="44" />
			<div>
				<span class="title">Vaultbot</span>
				<span class="subtitle mono">Genre graph</span>
			</div>
		</div>
		<nav>
			<a href="/">Graph</a>
		</nav>
	</div>
</header>

<main>
	{@render children()}
</main>

<footer>
	<span class="mono muted">vaultbot — genre graph</span>
</footer>

<style>
	header {
		border-bottom: 1px solid var(--border);
		background: var(--surface);
		padding: 0 2rem;
	}

	.header-inner {
		max-width: 1400px;
		margin: 0 auto;
		height: 64px;
		display: flex;
		align-items: center;
		justify-content: space-between;
	}

	.brand {
		display: flex;
		align-items: center;
		gap: 12px;
	}

	.title {
		display: block;
		font-size: 20px;
		font-weight: 600;
		letter-spacing: -0.02em;
		color: var(--text);
	}

	.subtitle {
		display: block;
		font-size: 11px;
		color: var(--text-muted);
	}

	nav a {
		color: var(--text-muted);
		font-size: 13px;
		transition: color 0.15s;
	}

	nav a:hover {
		color: var(--accent);
		text-decoration: none;
	}

	main {
		max-width: 1400px;
		margin: 0 auto;
		padding: 2rem;
	}

	footer {
		border-top: 1px solid var(--border);
		padding: 1rem 2rem;
		font-size: 12px;
		text-align: center;
	}
</style>
