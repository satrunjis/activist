import { describe, expect, it } from "vitest";

import { zoomLod } from "./zoomLod";

describe("zoomLod", () => {
  it("returns micro for zoom below 0.35", () => {
    expect(zoomLod(0.10)).toBe("micro");
    expect(zoomLod(0.34)).toBe("micro");
  });

  it("returns compact for zoom 0.35..0.65", () => {
    expect(zoomLod(0.35)).toBe("compact");
    expect(zoomLod(0.45)).toBe("compact");
    expect(zoomLod(0.65)).toBe("compact");
  });

  it("returns standard for zoom 0.66..1.10", () => {
    expect(zoomLod(0.66)).toBe("standard");
    expect(zoomLod(0.80)).toBe("standard");
    expect(zoomLod(1.10)).toBe("standard");
  });

  it("returns detail for zoom above 1.10", () => {
    expect(zoomLod(1.11)).toBe("detail");
    expect(zoomLod(1.50)).toBe("detail");
  });
});
