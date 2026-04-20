import { describe, expect, it } from "vitest";
import { fmtMonth, treemapColor } from "./stats";

describe("fmtMonth", () => {
	it("formats a YYYY-MM string to a short month/year label", () => {
		expect(fmtMonth("2024-01")).toBe("Jan 2024");
		expect(fmtMonth("2024-12")).toBe("Dec 2024");
	});

	it("handles single-digit month strings correctly", () => {
		expect(fmtMonth("2023-06")).toBe("Jun 2023");
	});

	it("returns consistent output for the same input", () => {
		expect(fmtMonth("2025-03")).toBe(fmtMonth("2025-03"));
	});
});

describe("treemapColor", () => {
	it("returns a valid rgba string", () => {
		expect(treemapColor(0)).toMatch(/^rgba\(124, 106, 247, [\d.]+\)$/);
		expect(treemapColor(1)).toMatch(/^rgba\(124, 106, 247, [\d.]+\)$/);
	});

	it("returns minimum alpha (0.28) at ratio 0", () => {
		const color = treemapColor(0);
		const alpha = Number(color.match(/rgba\([\d, ]+, ([\d.]+)\)/)?.[1]);
		expect(alpha).toBeCloseTo(0.28);
	});

	it("returns maximum alpha (0.90) at ratio 1", () => {
		const color = treemapColor(1);
		const alpha = Number(color.match(/rgba\([\d, ]+, ([\d.]+)\)/)?.[1]);
		expect(alpha).toBeCloseTo(0.9);
	});

	it("is monotonically increasing with ratio", () => {
		const a1 = Number(
			treemapColor(0.25).match(/rgba\([\d, ]+, ([\d.]+)\)/)?.[1],
		);
		const a2 = Number(
			treemapColor(0.75).match(/rgba\([\d, ]+, ([\d.]+)\)/)?.[1],
		);
		expect(a1).toBeLessThan(a2);
	});
});
