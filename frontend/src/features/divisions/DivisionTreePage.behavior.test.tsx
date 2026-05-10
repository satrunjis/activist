import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { beforeEach, describe, expect, it, vi } from "vitest";

import { DivisionsPage as DivisionTreePage } from "./DivisionsPage";
import type { DivisionChildrenResponse } from "../../shared/api/types";
import type { ApiRequestError } from "../../shared/api/errorPolicy";
import { ToastProvider } from "../../shared/ui/feedback";

vi.mock("../../shared/api/client", () => ({
  request: vi.fn()
}));

vi.mock("../positions/PositionListPanel", () => ({
  PositionListPanel: () => <div data-testid="position-list-mock" />
}));

import { request } from "../../shared/api/client";
const mockRequest = vi.mocked(request);

function renderDivisionTreePage() {
  render(
    <ToastProvider>
      <DivisionTreePage />
    </ToastProvider>
  );
}

const rootChildrenFixture: DivisionChildrenResponse = {
  items: [
    {
      id: "div-root",
      parent_id: "root",
      short_name: "Root",
      full_name: "Root Division",
      description: "Top level",
      is_archived: false,
      has_children: false,
      children_count: 0
    }
  ]
};

function accessError(): ApiRequestError {
  return {
    message: "Forbidden",
    code: "access.forbidden",
    status: 403,
    method: "POST",
    retryable: false
  };
}

describe("DivisionTreePage ACL boundary behavior", () => {
  beforeEach(() => {
    mockRequest.mockReset();
  });

  it("disables create and edit buttons and shows forbidden scope message after access error", async () => {
    const user = userEvent.setup();

    mockRequest.mockResolvedValueOnce(rootChildrenFixture);
    mockRequest.mockRejectedValueOnce(accessError());

    renderDivisionTreePage();

    // Wait for tree to load, then select a node so the detail panel opens
    // (use native click — userEvent pointer events are intercepted by ReactFlow)
    const divNode = await screen.findByTestId("division-node-div-root");
    divNode.click();

    // Panel opens with edit button; create button is in the toolbar
    expect(await screen.findByTestId("edit-division")).not.toBeDisabled();
    expect(screen.getByTestId("create-division")).not.toBeDisabled();

    // Open create form (panel already open from node selection)
    await user.click(screen.getByTestId("create-division"));
    await user.type(await screen.findByTestId("create-division-full-name"), "Test Division");
    await user.type(await screen.findByTestId("create-division-short-name"), "Test Div");
    await user.click(screen.getByTestId("create-division-submit"));

    expect(
      await screen.findByText(/операция запрещена в текущем контуре доступа\./i)
    ).toBeInTheDocument();

    expect(screen.getByTestId("create-division")).toBeDisabled();
    expect(screen.getByTestId("edit-division")).toBeDisabled();
  });

  it("eagerly loads all divisions including nested children on mount", async () => {
    mockRequest.mockImplementation(async (path: string) => {
      if (String(path).includes("?parent_id=root")) {
        return {
          items: [
            {
              id: "div-root",
              parent_id: "root",
              short_name: "Root",
              full_name: "Root Division",
              description: "Top level",
              is_archived: false,
              has_children: true,
              children_count: 1
            }
          ]
        };
      }
      if (String(path).includes("?parent_id=div-root")) {
        return {
          items: [
            {
              id: "child-1",
              parent_id: "div-root",
              short_name: "Child",
              full_name: "Child Division",
              description: "Nested",
              is_archived: false,
              has_children: false,
              children_count: 0
            }
          ]
        };
      }
      return { items: [] };
    });

    renderDivisionTreePage();

    // Both root division AND its child appear automatically without user interaction
    expect(await screen.findByTestId("division-node-div-root")).toBeInTheDocument();
    expect(await screen.findByTestId("division-node-child-1")).toBeInTheDocument();

    // Exactly one request for the nested level
    const childLoads = mockRequest.mock.calls.filter(([url]) =>
      String(url).includes("?parent_id=div-root")
    );
    expect(childLoads).toHaveLength(1);
  });
});
