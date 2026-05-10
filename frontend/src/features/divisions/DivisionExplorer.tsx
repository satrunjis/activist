import { useCallback, useEffect, useMemo, useReducer, useRef, useState } from "react";
import type { CSSProperties, ReactNode } from "react";
import { GitFork, MousePointerClick } from "lucide-react";

import { Button } from "../../components/ui/button";
import { request } from "../../shared/api/client";
import { adaptApiError } from "../../shared/api/errorAdapter";
import { ConfirmDialog, EmptyStateCard, InlineError, SectionSkeleton, SoftRefreshWrapper, useToast } from "../../shared/ui/feedback";
import { FullScreenPage, MetaBlock } from "../../shared/ui/layout";
import { ActionPanel, EntityCreateSurface, EntityEditSurface, PanelTabs } from "../../shared/ui/surfaces";
import type { PanelTab } from "../../shared/ui/surfaces";
import { AsyncStateView, type AsyncState } from "../../shared/ui/states/AsyncStateView";
import type {
  DivisionCreateRequest,
  DivisionEntity,
  DivisionPatchRequest,
  DivisionTreeNode,
  PositionItem,
  PositionMemberItem
} from "../../shared/api/types";
import { AssignMemberPanel } from "../positions/AssignMemberPanel";
import { PositionCreatePanel } from "../positions/PositionCreatePanel";
import { PositionListPanel } from "../positions/PositionListPanel";
import { PositionMembersList } from "../positions/PositionMembersList";
import { getPositionMembers } from "../positions/positionMembersApi";
import { DivisionEditForm } from "./DivisionEditForm";
import { fetchFullDivisionTree } from "./explorer/divisionTreeApi";
import { explorerReducer, initialExplorerState } from "./explorer/state/explorerReducer";
import { DivisionTreeViewport } from "./tree/DivisionTreeViewport";
import type { DivisionViewportNode } from "./tree/treeTypes";

type DivisionNodeCacheItem = Omit<DivisionTreeNode, "children">;

const ROOT_NODE: DivisionNodeCacheItem = {
  id: "root",
  short_name: "Корень",
  full_name: "Корень",
  description: "",
  is_archived: false,
  has_children: true,
  children_count: 0
};

const FORBIDDEN_SCOPE_MESSAGE = "Операция запрещена в текущем контуре доступа.";

const DIVISION_TABS: PanelTab[] = [
  { id: "info", label: "Инфо" },
  { id: "positions", label: "Должности" },
];

const FLOATING_PANEL_WIDTH = 380;
const FLOATING_PANEL_INSET = 8;

function toCachedNode(item: DivisionTreeNode): DivisionNodeCacheItem {
  return {
    id: item.id,
    parent_id: item.parent_id,
    short_name: item.short_name,
    full_name: item.full_name,
    description: item.description,
    regulation_url: item.regulation_url,
    media_links: item.media_links,
    is_archived: item.is_archived,
    has_children: item.has_children,
    children_count: item.children_count,
    positions_count: item.positions_count ?? 0,
    members_count: item.members_count ?? 0
  };
}

function singleLineTruncate(): CSSProperties {
  return {
    overflow: "hidden",
    textOverflow: "ellipsis",
    whiteSpace: "nowrap",
    minWidth: 0
  };
}

export function DivisionExplorer() {
  const [state, dispatch] = useReducer(explorerReducer, initialExplorerState);
  const [nodesById, setNodesById] = useState<Record<string, DivisionNodeCacheItem>>({});
  const [depthById, setDepthById] = useState<Record<string, number>>({});
  const [childrenByParent, setChildrenByParent] = useState<Record<string, string[]>>({});
  const [isBootstrapLoading, setBootstrapLoading] = useState(false);
  const [isRefreshing, setRefreshing] = useState(false);
  const [fitResetKey, setFitResetKey] = useState(0);
  const [errorMessage, setErrorMessage] = useState<string | null>(null);
  const [isSubmitting, setSubmitting] = useState(false);
  const [forbiddenMessage, setForbiddenMessage] = useState<string | null>(null);
  const [positionMembers, setPositionMembers] = useState<PositionMemberItem[]>([]);
  const [selectedPositionTitle, setSelectedPositionTitle] = useState<string | null>(null);
  const [archiveDialogOpen, setArchiveDialogOpen] = useState(false);
  const [activeTab, setActiveTab] = useState<"info" | "positions">("info");
  const memberRequestSeq = useRef(0);
  const pendingSelectRef = useRef<string | null>(null);
  const treeAbortControllerRef = useRef<AbortController | null>(null);
  const { showToast } = useToast();

  useEffect(() => {
    const id = pendingSelectRef.current;
    if (id && nodesById[id]) {
      pendingSelectRef.current = null;
      dispatch({ type: "clear-selection" });
      dispatch({ type: "select-division", divisionId: id });
    }
  }, [nodesById]);

  useEffect(() => {
    function handleEscapeKeyDown(event: KeyboardEvent) {
      if (event.key !== "Escape") {
        return;
      }
      dispatch({ type: "clear-selection" });
      setPositionMembers([]);
      setSelectedPositionTitle(null);
      setActiveTab("info");
    }

    window.addEventListener("keydown", handleEscapeKeyDown);
    return () => window.removeEventListener("keydown", handleEscapeKeyDown);
  }, []);

  const loadFullTree = useCallback(async (isRefresh = false, selectId?: string) => {
    if (treeAbortControllerRef.current) {
      treeAbortControllerRef.current.abort();
    }
    const controller = new AbortController();
    treeAbortControllerRef.current = controller;
    const { signal } = controller;

    if (isRefresh) {
      setRefreshing(true);
    } else {
      setFitResetKey((current) => current + 1);
      setBootstrapLoading(true);
    }
    setErrorMessage(null);

    try {
      const rootNode = await fetchFullDivisionTree(signal);
      const nodesMap: Record<string, DivisionNodeCacheItem> = {
        root: { ...ROOT_NODE }
      };
      const childrenMap: Record<string, string[]> = {};
      const depthMap: Record<string, number> = {};

      function walkNode(node: DivisionTreeNode, parentKey: string, depth: number) {
        nodesMap[node.id] = toCachedNode(node);
        depthMap[node.id] = depth;
        if (!childrenMap[parentKey]) childrenMap[parentKey] = [];
        childrenMap[parentKey].push(node.id);

        for (const child of node.children ?? []) {
          walkNode(child, node.id, depth + 1);
        }
      }

      walkNode(rootNode, "root", 0);

      const rootChildren = childrenMap["root"] ?? [];
      nodesMap["root"] = {
        ...ROOT_NODE,
        has_children: rootChildren.length > 0,
        children_count: rootChildren.length
      };

      if (selectId) {
        pendingSelectRef.current = selectId;
      }

      setNodesById(nodesMap);
      setChildrenByParent(childrenMap);
      setDepthById(depthMap);
      treeAbortControllerRef.current = null;
    } catch (error) {
      if (error instanceof DOMException && error.name === "AbortError") return;
      setErrorMessage(adaptApiError(error, "Не удалось загрузить оргструктуру.").message);
    } finally {
      if (isRefresh) {
        setRefreshing(false);
      } else {
        setBootstrapLoading(false);
      }
    }
  }, []);

  useEffect(() => {
    void loadFullTree();
    return () => { treeAbortControllerRef.current?.abort(); };
  }, [loadFullTree]);

  const viewportNodes = useMemo<DivisionViewportNode[]>(() => {
    const result: DivisionViewportNode[] = [];

    function visit(id: string, parentId?: string) {
      const node = nodesById[id];
      if (!node) {
        return;
      }

      result.push({
        id: node.id,
        parentId,
        depth: depthById[node.id] ?? 0,
        shortName: node.short_name,
        fullName: node.full_name,
        description: node.description,
        isArchived: node.is_archived,
        hasChildren: node.has_children,
        childrenCount: node.children_count,
        positionsCount: node.positions_count ?? 0,
        membersCount: node.members_count ?? 0,
        selected: state.selectedNodeId === node.id
      });

      for (const childId of childrenByParent[id] ?? []) {
        visit(childId, id);
      }
    }

    for (const childId of childrenByParent.root ?? []) {
      visit(childId, undefined);
    }
    return result;
  }, [childrenByParent, depthById, nodesById, state.selectedNodeId]);

  const selectedNode = useMemo(
    () => (state.selectedNodeId ? (nodesById[state.selectedNodeId] ?? null) : null),
    [nodesById, state.selectedNodeId]
  );

  const parentOptions = useMemo(
    () => Object.values(nodesById).map((item) => ({ id: item.id, label: item.full_name ?? item.short_name })),
    [nodesById]
  );

  const rootChildCount = (childrenByParent.root ?? []).length;

  async function handleCreate(payload: DivisionCreateRequest) {
    setSubmitting(true);
    try {
      const created = await request<DivisionEntity>("/api/v1/divisions", { method: "POST", body: payload });
      setForbiddenMessage(null);
      await loadFullTree(true, created.id);
      dispatch({ type: "save-success", divisionId: created.id });
    } catch (error) {
      const apiError = adaptApiError(error, "Не удалось создать подразделение.");
      if (apiError.kind === "forbidden") {
        setForbiddenMessage(FORBIDDEN_SCOPE_MESSAGE);
        dispatch({ type: "cancel" });
        return;
      }
      throw apiError;
    } finally {
      setSubmitting(false);
    }
  }

  async function handleEdit(payload: DivisionPatchRequest) {
    if (!selectedNode) {
      return;
    }
    setSubmitting(true);
    try {
      const updated = await request<DivisionEntity>(
        `/api/v1/divisions/${encodeURIComponent(selectedNode.id)}`,
        { method: "PATCH", body: payload }
      );
      setForbiddenMessage(null);
      await loadFullTree(true, updated.id);
      dispatch({ type: "save-success", divisionId: updated.id });
    } catch (error) {
      const apiError = adaptApiError(error, "Не удалось сохранить изменения подразделения.");
      if (apiError.kind === "forbidden") {
        setForbiddenMessage(FORBIDDEN_SCOPE_MESSAGE);
        dispatch({ type: "cancel" });
        return;
      }
      throw apiError;
    } finally {
      setSubmitting(false);
    }
  }

  async function handleArchiveSelectedDivision() {
    if (!selectedNode || selectedNode.id === "root") {
      return;
    }
    setSubmitting(true);
    try {
      await request(`/api/v1/divisions/${encodeURIComponent(selectedNode.id)}/archive`, { method: "POST" });
      setForbiddenMessage(null);
      await loadFullTree(true, selectedNode.parent_id ?? "root");
      showToast("success", "Подразделение архивировано");
    } catch (error) {
      const apiError = adaptApiError(error, "Не удалось архивировать подразделение.");
      if (apiError.kind === "forbidden") {
        setForbiddenMessage(FORBIDDEN_SCOPE_MESSAGE);
        return;
      }
      setErrorMessage(apiError.message);
    } finally {
      setSubmitting(false);
    }
  }

  async function handlePositionSelect(position: PositionItem) {
    dispatch({ type: "select-position", positionId: position.id });
    setSelectedPositionTitle(position.title);
    setPositionMembers([]);
    const seq = ++memberRequestSeq.current;
    dispatch({ type: "members-loading" });
    try {
      const response = await getPositionMembers(position.id);
      if (seq !== memberRequestSeq.current) {
        return;
      }
      setPositionMembers(response.items);
      dispatch({ type: "members-loaded" });
    } catch (error) {
      if (seq !== memberRequestSeq.current) {
        return;
      }
      dispatch({ type: "members-error", message: adaptApiError(error, "Не удалось загрузить участников должности.").message });
    }
  }

  const handleTreeNodeSelect = useCallback((id: string) => {
    dispatch({ type: "clear-selection" });
    dispatch({ type: "select-division", divisionId: id });
    setPositionMembers([]);
    setSelectedPositionTitle(null);
  }, []);

  const explorerState: AsyncState = isBootstrapLoading
    ? "loading"
    : errorMessage && !nodesById.root
      ? "error"
      : "success";

  const selectedNodeName = selectedNode?.full_name ?? selectedNode?.short_name;

  const panelHeader = useMemo<{ title: ReactNode; subtitle?: ReactNode }>(() => {
    switch (state.panelMode.kind) {
      case "empty":
        return {
          title: (
            <span style={{ ...singleLineTruncate(), fontSize: "var(--text-sm)", color: "var(--color-text-muted)" }}>
              Подразделение
            </span>
          )
        };
      case "division.view":
        return {
          title: (
            <span style={{ ...singleLineTruncate(), fontSize: "var(--text-sm)", color: "var(--color-text-primary)" }}>
              {selectedNodeName ?? "Подразделение"}
            </span>
          )
        };
      case "division.edit":
        return {
          title: "Редактирование",
          subtitle: selectedNodeName
        };
      case "division.create":
        return {
          title: "Новое подразделение",
          subtitle: selectedNode ? `в ${selectedNodeName}` : "Корневой уровень"
        };
      case "position.create":
        return {
          title: "Новая должность",
          subtitle: selectedNodeName
        };
      case "membership.create":
        return {
          title: "Назначение участника",
          subtitle: selectedNodeName
        };
      default:
        return { title: "Подразделение" };
    }
  }, [selectedNode, selectedNodeName, state.panelMode.kind]);

  const panelContent = (() => {
    switch (state.panelMode.kind) {
      case "empty":
        return (
          <div style={{ height: "100%", display: "flex", alignItems: "center", justifyContent: "center" }}>
            <EmptyStateCard
              body="Нажмите на узел дерева, чтобы увидеть его данные и действия."
              heading="Выберите подразделение"
              icon={<MousePointerClick size={32} />}
            />
          </div>
        );

      case "division.view":
        return (
          <>
            <PanelTabs
              activeTab={activeTab}
              onTabChange={(id) => setActiveTab(id as "info" | "positions")}
              tabs={DIVISION_TABS}
            />
            <div style={{ padding: "var(--space-4)" }}>
              {activeTab === "info" ? (
                <>
                  {selectedNode ? (
                    <>
                      <MetaBlock
                        fields={[
                          { label: "Полное название", value: selectedNode.full_name ?? selectedNode.short_name },
                          { label: "Краткое название", value: `@${selectedNode.short_name}` },
                          { label: "Описание", value: selectedNode.description || "—" },
                          { label: "Регламент", value: selectedNode.regulation_url ?? "—" },
                        ]}
                      />
                      {selectedNode.is_archived ? (
                        <span
                          style={{
                            display: "inline-flex",
                            marginTop: "var(--space-2)",
                            padding: "2px var(--space-2)",
                            borderRadius: "var(--radius-pill)",
                            background: "var(--color-warning-subtle)",
                            color: "var(--color-warning-text)",
                            fontSize: "var(--text-xs)"
                          }}
                        >
                          Архив
                        </span>
                      ) : null}
                      <div style={{ borderTop: "1px solid var(--color-border)", margin: "var(--space-4) 0" }} />
                      <div style={{ display: "flex", gap: "var(--space-2)" }}>
                        <Button
                          data-testid="edit-division"
                          disabled={selectedNode.id === "root" || Boolean(forbiddenMessage)}
                          onClick={() => dispatch({ type: "start-division-edit", divisionId: selectedNode.id })}
                          size="default"
                          type="button"
                          variant="outline"
                        >
                          Редактировать
                        </Button>
                        <Button
                          data-testid="archive-division"
                          disabled={selectedNode.id === "root" || selectedNode.is_archived || isSubmitting || Boolean(forbiddenMessage)}
                          onClick={() => setArchiveDialogOpen(true)}
                          size="default"
                          type="button"
                          variant="destructive"
                        >
                          Архивировать
                        </Button>
                      </div>
                      <InlineError message={forbiddenMessage} />
                      <InlineError message={errorMessage} />
                    </>
                  ) : null}
                </>
              ) : (
                <>
                  {selectedNode && selectedNode.id !== "root" ? (
                    <>
                      <PositionListPanel
                        divisionId={selectedNode.id}
                        onSelectPosition={handlePositionSelect}
                        onStartCreatePosition={(divisionId) => {
                          setActiveTab("positions");
                          dispatch({ type: "cancel" });
                          dispatch({ type: "start-position-create", divisionId });
                        }}
                        onStartMembershipCreate={(divisionId, positionId) => {
                          setActiveTab("positions");
                          dispatch({ type: "cancel" });
                          dispatch({ type: "start-membership-create", divisionId, positionId });
                        }}
                        selectedPositionId={state.selectedPositionId}
                      />
                      {state.selectedPositionId ? (
                        <>
                          <div
                            style={{
                              marginTop: "var(--space-4)",
                              marginBottom: "var(--space-2)",
                              fontSize: "var(--text-xs)",
                              fontWeight: "var(--weight-semibold)",
                              color: "var(--color-text-muted)",
                              textTransform: "uppercase",
                              letterSpacing: "0.06em"
                            }}
                          >
                            {selectedPositionTitle ? `Участники: ${selectedPositionTitle}` : "Участники"}
                          </div>
                          <PositionMembersList
                            errorMessage={state.memberError}
                            items={positionMembers}
                            positionId={state.selectedPositionId}
                            state={state.memberLoadState}
                          />
                        </>
                      ) : null}
                    </>
                  ) : null}
                </>
              )}
            </div>
          </>
        );

      case "division.edit":
        return (
          <EntityEditSurface onClose={() => dispatch({ type: "cancel" })} showTitle={false}>
            <DivisionEditForm
              isSubmitting={isSubmitting}
              mode="edit"
              node={selectedNode!}
              onCancel={() => dispatch({ type: "cancel" })}
              onCreate={handleCreate}
              onEdit={handleEdit}
              parentOptions={parentOptions}
            />
          </EntityEditSurface>
        );

      case "division.create":
        return (
          <EntityCreateSurface onClose={() => dispatch({ type: "cancel" })} showTitle={false}>
            <DivisionEditForm
              defaultParentId={state.panelMode.parentId ?? undefined}
              isSubmitting={isSubmitting}
              mode="create"
              onCancel={() => dispatch({ type: "cancel" })}
              onCreate={handleCreate}
              onEdit={handleEdit}
              parentOptions={parentOptions}
            />
          </EntityCreateSurface>
        );

      case "position.create":
        return (
          <PositionCreatePanel
            divisionId={state.panelMode.divisionId}
            onSuccess={() => dispatch({ type: "save-success", divisionId: state.panelMode.divisionId })}
            onClose={() => dispatch({ type: "cancel" })}
          />
        );

      case "membership.create":
        return state.panelMode.positionId ? (
          <AssignMemberPanel
            positionId={state.panelMode.positionId}
            positionTitle={selectedPositionTitle ?? state.panelMode.positionId}
            onSuccess={() => dispatch({ type: "save-success", divisionId: state.panelMode.divisionId })}
            onClose={() => dispatch({ type: "cancel" })}
          />
        ) : (
          <EmptyStateCard
            body=""
            heading="Сначала выберите должность"
          />
        );

      default:
        return null;
    }
  })();

  return (
    <AsyncStateView
      data={true}
      errorMessage={errorMessage}
      loadingView={(
        <FullScreenPage data-testid="division-tree">
          <div style={{ padding: "32px", width: "100%" }}>
            <SectionSkeleton rows={5} withHeader />
          </div>
        </FullScreenPage>
      )}
      onRetry={() => void loadFullTree()}
      state={explorerState}
    >
      {() => (
        <>
          <FullScreenPage data-testid="division-tree">
            <div style={{ position: "relative", flex: 1, minWidth: 0, minHeight: 0, overflow: "hidden" }}>
              <div style={{ position: "absolute", inset: 0, overflow: "hidden" }}>
                <SoftRefreshWrapper isRefreshing={isRefreshing}>
                  {rootChildCount === 0 ? (
                    <div
                      style={{
                        position: "absolute",
                        inset: 0,
                        display: "flex",
                        alignItems: "center",
                        justifyContent: "center",
                        padding: "var(--space-6)"
                      }}
                    >
                      <EmptyStateCard
                        body="Создайте первое подразделение, чтобы начать строить оргструктуру."
                        cta={!forbiddenMessage
                          ? {
                            label: "Создать подразделение",
                            onClick: () => {
                              dispatch({ type: "clear-selection" });
                              dispatch({ type: "start-division-create", parentId: null });
                              setPositionMembers([]);
                            }
                          }
                          : undefined}
                        heading="Оргструктура пуста"
                        icon={<GitFork size={40} />}
                      />
                    </div>
                  ) : (
                    <DivisionTreeViewport
                      fitResetKey={fitResetKey}
                      isRefreshing={isRefreshing}
                      nodes={viewportNodes}
                      onSelect={handleTreeNodeSelect}
                      panelWidth={FLOATING_PANEL_WIDTH}
                    />
                  )}
                </SoftRefreshWrapper>
              </div>

              {(() => {
                const isFormMode = (
                  state.panelMode.kind === "division.edit" ||
                  state.panelMode.kind === "division.create" ||
                  state.panelMode.kind === "position.create" ||
                  state.panelMode.kind === "membership.create"
                );
                return (
                  <div
                    style={{
                      position: "absolute",
                      bottom: FLOATING_PANEL_INSET,
                      left: FLOATING_PANEL_INSET,
                      display: "flex",
                      alignItems: "center",
                      gap: "var(--space-2)",
                      zIndex: 1002
                    }}
                  >
                    <Button
                      data-testid="refresh-division-tree"
                      disabled={isRefreshing || isFormMode}
                      onClick={() => void loadFullTree(true, state.selectedNodeId ?? undefined)}
                      size="default"
                      style={{ minWidth: 92 }}
                      type="button"
                      variant="ghost"
                    >
                      Обновить
                    </Button>
                    <Button
                      data-testid="create-division"
                      disabled={Boolean(forbiddenMessage) || isFormMode}
                      onClick={() => {
                        dispatch({ type: "cancel" });
                        dispatch({ type: "start-division-create", parentId: selectedNode?.id ?? null });
                        setPositionMembers([]);
                      }}
                      size="default"
                      style={{ minWidth: 136 }}
                      type="button"
                      variant="primary"
                    >
                      + Подразделение
                    </Button>
                  </div>
                );
              })()}

              <div
                data-testid="division-right-panel"
                style={{
                  position: "absolute",
                  top: FLOATING_PANEL_INSET,
                  right: FLOATING_PANEL_INSET,
                  bottom: FLOATING_PANEL_INSET,
                  width: FLOATING_PANEL_WIDTH,
                  display: "flex",
                  flexDirection: "column",
                  background: "transparent",
                  minHeight: 0,
                  zIndex: 10
                }}
              >
                <ActionPanel
                  subtitle={panelHeader.subtitle}
                  title={panelHeader.title}
                  width={FLOATING_PANEL_WIDTH}
                >
                  {panelContent}
                </ActionPanel>
              </div>
            </div>
          </FullScreenPage>

          <ConfirmDialog
            body={
              selectedNode
                ? `Подразделение «${selectedNode.full_name ?? selectedNode.short_name}» и все его дочерние подразделения будут переведены в архив. Это действие можно отменить позже.`
                : ""
            }
            confirmLabel="Архивировать"
            destructive
            onCancel={() => setArchiveDialogOpen(false)}
            onConfirm={() => {
              setArchiveDialogOpen(false);
              void handleArchiveSelectedDivision();
            }}
            open={archiveDialogOpen}
            title="Архивировать подразделение?"
          />
        </>
      )}
    </AsyncStateView>
  );
}
