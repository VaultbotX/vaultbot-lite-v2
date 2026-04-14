import { serve } from '@hono/node-server';
import { Hono } from 'hono';
import qs from 'querystring';
import { generateRandomString } from './generate-random-string.js';
import dotenv from 'dotenv';

const configResult = dotenv.config();
if (configResult.error) {
	console.error('Failed to load .env file', configResult.error);
	process.exit(1);
}

const SPOTIFY_CLIENT_ID = process.env.SPOTIFY_CLIENT_ID;
if (!SPOTIFY_CLIENT_ID) {
	console.error('SPOTIFY_CLIENT_ID is not set');
	process.exit(1);
}

const SPOTIFY_REDIRECT_URI = process.env.SPOTIFY_REDIRECT_URI;
if (!SPOTIFY_REDIRECT_URI) {
	console.error('SPOTIFY_REDIRECT_URI is not set');
	process.exit(1);
}

const SPOTIFY_CLIENT_SECRET = process.env.SPOTIFY_CLIENT_SECRET;
if (!SPOTIFY_CLIENT_SECRET) {
	console.error('SPOTIFY_CLIENT_SECRET is not set');
	process.exit(1);
}

const port = 8888;
const app = new Hono();

// https://developer.spotify.com/documentation/web-api/tutorials/code-flow

app.get('/login', (c) => {
	const state = generateRandomString(16);
	console.log(`Generated state for use in following requests: ${state}`);

	// https://developer.spotify.com/documentation/web-api/concepts/scopes
	const scope = 'playlist-modify-public playlist-modify-private playlist-read-private playlist-read-collaborative';

	const url = 'https://accounts.spotify.com/authorize?' + qs.stringify({
		response_type: 'code',
		client_id: SPOTIFY_CLIENT_ID,
		scope,
		redirect_uri: SPOTIFY_REDIRECT_URI,
		state
	});

	console.log(`Redirecting to ${url}`);

	return c.redirect(url);
});

app.get('/callback', async (c) => {
	const { code, state } = c.req.query();
	if (!code) {
		return c.text('No code provided', 400);
	}
	if (!state) {
		return c.text('No state provided', 400);
	}

	console.log({
		code,
		state
	});

	const base64Auth = Buffer.from(`${SPOTIFY_CLIENT_ID}:${SPOTIFY_CLIENT_SECRET}`).toString('base64');

	const requestBody = {
		code: code,
		redirect_uri: SPOTIFY_REDIRECT_URI,
		grant_type: 'authorization_code'
	};

	console.log('Requesting access token');

	const res = await fetch('https://accounts.spotify.com/api/token', {
		method: 'POST',
		headers: {
			'content-type': 'application/x-www-form-urlencoded',
			'Authorization': 'Basic ' + base64Auth
		},
		body: new URLSearchParams(requestBody)
	});

	if (!res.ok) {
		console.error('Failed to get access token', res.status, await res.text());
		return c.text('Failed to get access token', 422);
	}

	const responseBody = await res.json();

	console.log(responseBody);

	return c.json(responseBody, 200);
});

console.log(`Server is running on port ${port}`);

serve({
	fetch: app.fetch,
	port
});
