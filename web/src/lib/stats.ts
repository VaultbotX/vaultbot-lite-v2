// Formats a "YYYY-MM" string to a short locale date, e.g. "Jan 2024"
export function fmtMonth(ym: string): string {
	const [year, month] = ym.split("-");
	return new Date(Number(year), Number(month) - 1).toLocaleDateString("en-US", {
		month: "short",
		year: "numeric",
	});
}

// Maps a [0,1] ratio to the purple accent colour used in the charts
export function treemapColor(ratio: number): string {
	const alpha = 0.28 + ratio * 0.62;
	return `rgba(124, 106, 247, ${alpha.toFixed(2)})`;
}
