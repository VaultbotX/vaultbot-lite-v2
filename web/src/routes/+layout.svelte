<script lang="ts">
import "../app.css";
import { page } from "$app/state";

let { children } = $props();

const fullBleed = $derived(page.url.pathname.startsWith("/galaxy"));
</script>

<svelte:head>
	<title>Vaultbot</title>
</svelte:head>

<header>
	<div class="header-inner">
		<a class="brand" href="/">
			<img src="/vaultbot.png" alt="Vaultbot logo" width="44" height="44" />
			<div>
				<span class="title">Vaultbot</span>
				<span class="subtitle mono">Playlist analytics</span>
			</div>
		</a>
		<nav>
			<a href="/">Home</a>
			<a href="/galaxy">Playlist Galaxy</a>
		</nav>
	</div>
</header>

<main class:full-bleed={fullBleed}>
	{@render children()}
</main>

<footer>
	<span class="mono muted">vaultbot — data refreshed every 6 hours</span>
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
		color: inherit;
		text-decoration: none;
	}

	.brand:hover {
		text-decoration: none;
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

	nav {
		display: flex;
		gap: 1.5rem;
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

	main.full-bleed {
		max-width: none;
	}

	footer {
		border-top: 1px solid var(--border);
		padding: 1rem 2rem;
		font-size: 12px;
		text-align: center;
	}
</style>
