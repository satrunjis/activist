export type DivisionPanelMode =
  | { kind: "empty" }
  | { kind: "division.view"; divisionId: string }
  | { kind: "division.edit"; divisionId: string }
  | { kind: "division.create"; parentId: string | null }
  | { kind: "position.create"; divisionId: string }
  | { kind: "membership.create"; divisionId: string; positionId: string };
export type ExplorerMemberLoadState = "idle" | "loading" | "loaded" | "error";

export type ExplorerState = {
  selectedNodeId: string | null;
  selectedPositionId: string | null;
  panelMode: DivisionPanelMode;
  memberLoadState: ExplorerMemberLoadState;
  memberError: string | null;
};

export type ExplorerAction =
  | { type: "select-division"; divisionId: string }
  | { type: "start-division-edit"; divisionId: string }
  | { type: "start-division-create"; parentId: string | null }
  | { type: "start-position-create"; divisionId: string }
  | { type: "start-membership-create"; divisionId: string; positionId: string }
  | { type: "cancel" }
  | { type: "save-success"; divisionId: string }
  | { type: "clear-selection" }
  | { type: "select-position"; positionId: string }
  | { type: "members-loading" }
  | { type: "members-loaded" }
  | { type: "members-error"; message: string };

export const initialExplorerState: ExplorerState = {
  selectedNodeId: null,
  selectedPositionId: null,
  panelMode: { kind: "empty" },
  memberLoadState: "idle",
  memberError: null
};

export function explorerReducer(state: ExplorerState, action: ExplorerAction): ExplorerState {
  switch (action.type) {
    case "select-division":
      if (state.panelMode.kind !== "empty") {
        return state;
      }
      return {
        ...state,
        selectedNodeId: action.divisionId,
        selectedPositionId: null,
        panelMode: { kind: "division.view", divisionId: action.divisionId },
        memberLoadState: "idle",
        memberError: null
      };
    case "start-division-edit":
      if (state.panelMode.kind !== "division.view" || state.panelMode.divisionId !== action.divisionId) {
        return state;
      }
      return {
        ...state,
        panelMode: { kind: "division.edit", divisionId: action.divisionId }
      };
    case "start-division-create":
      if (state.panelMode.kind !== "empty" && state.panelMode.kind !== "division.view") {
        return state;
      }
      return {
        ...state,
        panelMode: { kind: "division.create", parentId: action.parentId }
      };
    case "start-position-create":
      if (state.panelMode.kind !== "division.view" || state.panelMode.divisionId !== action.divisionId) {
        return state;
      }
      return {
        ...state,
        panelMode: { kind: "position.create", divisionId: action.divisionId }
      };
    case "start-membership-create":
      if (state.panelMode.kind !== "division.view" || state.panelMode.divisionId !== action.divisionId) {
        return state;
      }
      return {
        ...state,
        selectedPositionId: action.positionId,
        panelMode: {
          kind: "membership.create",
          divisionId: action.divisionId,
          positionId: action.positionId
        }
      };
    case "cancel":
      switch (state.panelMode.kind) {
        case "division.edit":
          return {
            ...state,
            panelMode: { kind: "division.view", divisionId: state.panelMode.divisionId }
          };
        case "division.create":
          if (state.panelMode.parentId) {
            return {
              ...state,
              selectedNodeId: state.panelMode.parentId,
              panelMode: { kind: "division.view", divisionId: state.panelMode.parentId }
            };
          }
          return {
            ...state,
            selectedNodeId: null,
            selectedPositionId: null,
            panelMode: { kind: "empty" },
            memberLoadState: "idle",
            memberError: null
          };
        case "position.create":
          return {
            ...state,
            selectedNodeId: state.panelMode.divisionId,
            panelMode: { kind: "division.view", divisionId: state.panelMode.divisionId }
          };
        case "membership.create":
          return {
            ...state,
            selectedNodeId: state.panelMode.divisionId,
            panelMode: { kind: "division.view", divisionId: state.panelMode.divisionId }
          };
        default:
          return state;
      }
    case "save-success":
      switch (state.panelMode.kind) {
        case "division.edit":
        case "division.create":
        case "position.create":
        case "membership.create":
          return {
            ...state,
            selectedNodeId: action.divisionId,
            panelMode: { kind: "division.view", divisionId: action.divisionId }
          };
        default:
          return state;
      }
    case "clear-selection":
      if (state.panelMode.kind === "empty") {
        return state;
      }
      return {
        ...state,
        selectedNodeId: null,
        selectedPositionId: null,
        panelMode: { kind: "empty" },
        memberLoadState: "idle",
        memberError: null
      };
    case "select-position":
      return {
        ...state,
        selectedPositionId: action.positionId,
        memberLoadState: "idle",
        memberError: null
      };
    case "members-loading":
      return {
        ...state,
        memberLoadState: "loading",
        memberError: null
      };
    case "members-loaded":
      return {
        ...state,
        memberLoadState: "loaded",
        memberError: null
      };
    case "members-error":
      return {
        ...state,
        memberLoadState: "error",
        memberError: action.message
      };
    default:
      return state;
  }
}
