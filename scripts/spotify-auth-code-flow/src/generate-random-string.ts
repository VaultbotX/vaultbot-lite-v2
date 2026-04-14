export function generateRandomString(length: number = 16) {
	return [...Array(length)]
		.map(() => (~~(Math.random() * 36)).toString(36))
		.join('');
}
