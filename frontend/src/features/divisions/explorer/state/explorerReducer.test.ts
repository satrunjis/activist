import { describe, expect, it } from "vitest";

import { explorerReducer, initialExplorerState } from "./explorerReducer";

describe("explorerReducer", () => {
  it("opens panel and binds owner on node select", () => {
    const next = explorerReducer(initialExplorerState, { type: "select-division", divisionId: "div-1" });

    expect(next.selectedNodeId).toBe("div-1");
    expect(next.panelMode).toEqual({ kind: "division.view", divisionId: "div-1" });
    expect(next.selectedPositionId).toBeNull();
  });

  it("position select resets stale member load state", () => {
    const withError = {
      ...initialExplorerState,
      selectedNodeId: "div-1",
      panelMode: { kind: "division.view", divisionId: "div-1" } as const,
      memberLoadState: "error" as const,
      memberError: "failed before"
    };
    const next = explorerReducer(withError, { type: "select-position", positionId: "pos-1" });

    expect(next.selectedPositionId).toBe("pos-1");
    expect(next.memberLoadState).toBe("idle");
    expect(next.memberError).toBeNull();
  });

  it("cancel from edit mode returns to selected division view", () => {
    const selected = explorerReducer(initialExplorerState, { type: "select-division", divisionId: "div-1" });
    const editing = explorerReducer(selected, { type: "start-division-edit", divisionId: "div-1" });
    const canceled = explorerReducer(editing, { type: "cancel" });

    expect(canceled.selectedNodeId).toBe("div-1");
    expect(canceled.panelMode).toEqual({ kind: "division.view", divisionId: "div-1" });
  });
});
