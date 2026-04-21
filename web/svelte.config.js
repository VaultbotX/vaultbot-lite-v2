import adapter from "@sveltejs/adapter-cloudflare";
import { vitePreprocess } from "@sveltejs/vite-plugin-svelte";

/** @type {import('@sveltejs/kit').Config} */
const config = {
	preprocess: vitePreprocess(),
	kit: {
		adapter: adapter(),
	},
	vitePlugin: {
		inspector: {
      showToggleButton: 'always',
      toggleButtonPos: 'bottom-right'
    }
	},
};

export default config;
