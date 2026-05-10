package division

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	appauth "activist-base/src/application/auth"
	appdivision "activist-base/src/application/division"
	appeventlog "activist-base/src/application/eventlog"
	domaindivision "activist-base/src/domain/division"
	domaineventlog "activist-base/src/domain/eventlog"
	domainmembership "activist-base/src/domain/membership"
	"activist-base/src/domain/role"
	"activist-base/src/domain/shared"
	postgres "activist-base/src/repository/postgres"
	httpsession "activist-base/src/transport/http/session"

	"github.com/go-chi/chi/v5"
	chiMiddleware "github.com/go-chi/chi/v5/middleware"
)

const defaultTreeDepth = 8

type Authenticator interface {
	Authenticate(next http.Handler) http.Handler
}

type Repository interface {
	GetRoot(ctx context.Context) (domaindivision.Division, error)
	GetByID(ctx context.Context, id shared.DivisionID) (domaindivision.Division, error)
	Create(ctx context.Context, division domaindivision.Division) (domaindivision.Division, error)
	Update(ctx context.Context, division domaindivision.Division) (domaindivision.Division, error)
	ArchiveCascadeWithTx(ctx context.Context, tx appeventlog.Transaction, id shared.DivisionID) (domaindivision.Division, int, int, bool, error)
	GetAncestorIDs(ctx context.Context, id shared.DivisionID) ([]shared.DivisionID, error)
	ListTree(ctx context.Context, includeArchived bool) ([]domaindivision.Division, error)
	ListChildren(ctx context.Context, parentID shared.DivisionID) ([]postgres.DivisionChildItem, error)
	ListRootChildren(ctx context.Context) ([]postgres.DivisionChildItem, error)
}

type AuthorizationService interface {
	ResolveForDivision(ctx context.Context, actorID shared.UserID, targetDivisionID shared.DivisionID) (domainmembership.EffectivePermissions, error)
	ResolveGlobal(ctx context.Context, actorID shared.UserID) (domainmembership.EffectivePermissions, error)
}

type Handler struct {
	repo     Repository
	authz    AuthorizationService
	commands *appeventlog.CommandService
	nowFn    func() time.Time
}

func NewHandler(repo Repository, authz AuthorizationService, commands ...*appeventlog.CommandService) *Handler {
	var commandService *appeventlog.CommandService
	if len(commands) > 0 {
		commandService = commands[0]
	}
	return &Handler{
		repo:     repo,
		authz:    authz,
		commands: commandService,
		nowFn: func() time.Time {
			return time.Now().UTC()
		},
	}
}

func RegisterRoutes(r chi.Router, handler *Handler, authMiddleware Authenticator) {
	if r == nil || handler == nil {
		return
	}

	r.Route("/divisions", func(divisions chi.Router) {
		if authMiddleware != nil {
			divisions.Use(authMiddleware.Authenticate)
		}
		divisions.Get("/", handler.GetChildren)
		divisions.Post("/", handler.Create)
		divisions.Patch("/{division_id}", handler.Patch)
		divisions.Post("/{division_id}/archive", handler.Archive)
		divisions.Get("/tree", handler.GetTree)
	})
}

// NewTestRouter mounts division endpoints without auth middleware for handler tests.
func NewTestRouter(handler *Handler) http.Handler {
	r := chi.NewRouter()
	r.Use(chiMiddleware.RequestID)
	r.Route("/api/v1", func(api chi.Router) {
		RegisterRoutes(api, handler, nil)
	})
	return r
}

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	actor, ok := httpsession.ActorFromRequest(r)
	if !ok {
		writeError(w, r, appauth.ErrInvalidSession)
		return
	}

	request, err := decodeCreateRequest(r)
	if err != nil {
		writeError(w, r, err)
		return
	}

	var (
		parent         *domaindivision.Division
		existingRootID *shared.DivisionID
	)
	if request.ParentID != nil {
		parentID := shared.DivisionID(strings.TrimSpace(*request.ParentID))
		if parentID == "" {
			writeError(w, r, shared.ErrEmptyID)
			return
		}

		parentDivision, err := h.repo.GetByID(r.Context(), parentID)
		if err != nil {
			writeDivisionError(w, r, err)
			return
		}
		parent = &parentDivision
	} else {
		root, err := h.repo.GetRoot(r.Context())
		if err == nil {
			rootID := root.ID
			existingRootID = &rootID
		} else if !errors.Is(err, sql.ErrNoRows) {
			writeDivisionError(w, r, err)
			return
		}
	}

	access, err := h.resolveCreateAccess(r.Context(), actor.UserID, request.ParentID)
	if err != nil {
		writeDivisionError(w, r, err)
		return
	}

	result, err := appdivision.Create(appdivision.CreateInput{
		ActorID:        actor.UserID,
		ID:             shared.DivisionID(newDivisionID()),
		Parent:         parent,
		ExistingRootID: existingRootID,
		ShortName:      request.ShortName,
		FullName:       request.FullName,
		Description:    request.Description,
		RegulationURL:  request.RegulationURL,
		MediaLinks:     request.MediaLinks,
		Access:         access,
		Now:            h.nowFn(),
	})
	if err != nil {
		writeDivisionError(w, r, err)
		return
	}

	persisted, err := h.repo.Create(r.Context(), result.Division)
	if err != nil {
		writeDivisionError(w, r, err)
		return
	}

	writeJSON(w, http.StatusCreated, divisionDTO(persisted))
}

func newDivisionID() string {
	buf := make([]byte, 16)
	if _, err := rand.Read(buf); err != nil {
		return strconv.FormatInt(time.Now().UnixNano(), 16)
	}
	return hex.EncodeToString(buf)
}

func (h *Handler) Patch(w http.ResponseWriter, r *http.Request) {
	actor, ok := httpsession.ActorFromRequest(r)
	if !ok {
		writeError(w, r, appauth.ErrInvalidSession)
		return
	}

	divisionID := shared.DivisionID(strings.TrimSpace(chi.URLParam(r, "division_id")))
	if divisionID == "" {
		writeError(w, r, shared.ErrEmptyID)
		return
	}

	existing, err := h.repo.GetByID(r.Context(), divisionID)
	if err != nil {
		writeDivisionError(w, r, err)
		return
	}

	request, err := decodePatchRequest(r)
	if err != nil {
		writeError(w, r, err)
		return
	}

	access, err := h.resolveDivisionAccess(r.Context(), actor.UserID, divisionID)
	if err != nil {
		writeDivisionError(w, r, err)
		return
	}

	editInput := appdivision.EditInput{
		ActorID:  actor.UserID,
		Existing: existing,
		Patch:    toDivisionPatch(request),
		Access:   access,
		Now:      h.nowFn(),
	}
	if request.ParentID != nil {
		parentID := shared.DivisionID(strings.TrimSpace(*request.ParentID))
		if parentID == "" {
			writeError(w, r, shared.ErrEmptyID)
			return
		}

		parent, err := h.repo.GetByID(r.Context(), parentID)
		if err != nil {
			writeDivisionError(w, r, err)
			return
		}
		ancestorIDs, err := h.repo.GetAncestorIDs(r.Context(), parentID)
		if err != nil {
			writeDivisionError(w, r, err)
			return
		}
		editInput.ReassignParent = true
		editInput.NewParent = &parent
		editInput.NewParentAncestors = ancestorIDs
	}

	result, err := appdivision.Edit(editInput)
	if err != nil {
		writeDivisionError(w, r, err)
		return
	}

	persisted, err := h.repo.Update(r.Context(), result.Division)
	if err != nil {
		writeDivisionError(w, r, err)
		return
	}

	writeJSON(w, http.StatusOK, divisionDTO(persisted))
}

func (h *Handler) GetTree(w http.ResponseWriter, r *http.Request) {
	actor, ok := httpsession.ActorFromRequest(r)
	if !ok {
		writeError(w, r, appauth.ErrInvalidSession)
		return
	}

	treeDepth := defaultTreeDepth
	if rawDepth := strings.TrimSpace(r.URL.Query().Get("depth")); rawDepth != "" {
		parsedDepth, err := strconv.Atoi(rawDepth)
		if err != nil || parsedDepth <= 0 {
			writeError(w, r, &shared.Error{
				Code:    "validation.tree_depth",
				Message: "depth must be a positive integer",
			})
			return
		}
		treeDepth = parsedDepth
	}

	includeArchived := r.URL.Query().Get("include_archived") == "true"

	divisions, err := h.repo.ListTree(r.Context(), includeArchived)
	if err != nil {
		writeDivisionError(w, r, err)
		return
	}
	rootNode, err := toTreeNode(divisions)
	if err != nil {
		writeDivisionError(w, r, err)
		return
	}

	access, err := h.resolveGlobalAccess(r.Context(), actor.UserID)
	if err != nil {
		writeDivisionError(w, r, err)
		return
	}

	result, err := appdivision.GetTree(appdivision.TreeInput{
		Root:        rootNode,
		Depth:       treeDepth,
		CanViewCard: access.Has(role.CanEditDivision, false),
	})
	if err != nil {
		writeDivisionError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, toDivisionResponse(result))
}

// GetChildren handles GET /api/v1/divisions with optional parent_id query param.
//
// Semantics:
//   - parent_id omitted or "root" => root-level divisions
//   - parent_id=<non-empty id>    => direct children of that division
//
// Response is a JSON list of child items with has_children and children_count
// so the UI can show expand affordances without fetching the full tree.
func (h *Handler) GetChildren(w http.ResponseWriter, r *http.Request) {
	actor, ok := httpsession.ActorFromRequest(r)
	if !ok {
		writeError(w, r, appauth.ErrInvalidSession)
		return
	}

	rawParent := strings.TrimSpace(r.URL.Query().Get("parent_id"))

	var items []postgres.DivisionChildItem
	var fetchErr error

	if rawParent == "" || rawParent == "root" {
		// root mode: return top-level divisions
		items, fetchErr = h.repo.ListRootChildren(r.Context())
	} else {
		parentID := shared.DivisionID(rawParent)
		if parentID == "" {
			writeError(w, r, shared.ErrEmptyID)
			return
		}
		items, fetchErr = h.repo.ListChildren(r.Context(), parentID)
	}

	if fetchErr != nil {
		writeDivisionError(w, r, fetchErr)
		return
	}

	// Resolve global access to determine card visibility — consistent with GetTree.
	access, err := h.resolveGlobalAccess(r.Context(), actor.UserID)
	if err != nil {
		writeDivisionError(w, r, err)
		return
	}
	canViewCard := access.Has(role.CanEditDivision, false)

	dtos := make([]childItemDTO, 0, len(items))
	for _, item := range items {
		dtos = append(dtos, toChildItemDTO(item, canViewCard))
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": dtos})
}

func (h *Handler) Archive(w http.ResponseWriter, r *http.Request) {
	actor, ok := httpsession.ActorFromRequest(r)
	if !ok {
		writeError(w, r, appauth.ErrInvalidSession)
		return
	}

	divisionID := shared.DivisionID(strings.TrimSpace(chi.URLParam(r, "division_id")))
	if divisionID == "" {
		writeError(w, r, shared.ErrEmptyID)
		return
	}

	existing, err := h.repo.GetByID(r.Context(), divisionID)
	if err != nil {
		writeDivisionError(w, r, err)
		return
	}

	access, err := h.resolveDivisionAccess(r.Context(), actor.UserID, divisionID)
	if err != nil {
		writeDivisionError(w, r, err)
		return
	}

	if _, err := appdivision.Archive(appdivision.ArchiveInput{
		ActorID:                 actor.UserID,
		Division:                existing,
		ArchivedPositionsCount:  0,
		RemovedMembershipsCount: 0,
		Access:                  access,
		Now:                     h.nowFn(),
	}); err != nil {
		if errors.Is(err, domaindivision.ErrArchived) {
			writeJSON(w, http.StatusOK, divisionDTO(existing))
			return
		}
		writeDivisionError(w, r, err)
		return
	}

	if h.commands == nil {
		writeError(w, r, errors.New("division command service is not configured"))
		return
	}

	var persisted domaindivision.Division
	err = h.commands.ExecuteWithEvent(r.Context(), func(ctx context.Context, tx appeventlog.Transaction) (domaineventlog.Entry, error) {
		archivedDivision, removedMemberships, archivedPositions, didArchive, txErr := h.repo.ArchiveCascadeWithTx(ctx, tx, divisionID)
		if txErr != nil {
			return domaineventlog.Entry{}, txErr
		}
		if !didArchive {
			persisted = archivedDivision
			return domaineventlog.Entry{}, errDivisionArchiveNoop
		}

		result, txErr := appdivision.Archive(appdivision.ArchiveInput{
			ActorID:                 actor.UserID,
			Division:                existing,
			ArchivedPositionsCount:  archivedPositions,
			RemovedMembershipsCount: removedMemberships,
			Access:                  access,
			Now:                     h.nowFn(),
		})
		if txErr != nil {
			return domaineventlog.Entry{}, txErr
		}
		persisted = archivedDivision
		return result.Event, nil
	})
	if errors.Is(err, errDivisionArchiveNoop) {
		writeJSON(w, http.StatusOK, divisionDTO(persisted))
		return
	}
	if err != nil {
		writeDivisionError(w, r, err)
		return
	}

	writeJSON(w, http.StatusOK, divisionDTO(persisted))
}

func (h *Handler) resolveDivisionAccess(ctx context.Context, actorID shared.UserID, divisionID shared.DivisionID) (domainmembership.EffectivePermissions, error) {
	if h.authz == nil {
		return domainmembership.EffectivePermissions{}, shared.ErrForbidden
	}
	return h.authz.ResolveForDivision(ctx, actorID, divisionID)
}

func (h *Handler) resolveGlobalAccess(ctx context.Context, actorID shared.UserID) (domainmembership.EffectivePermissions, error) {
	if h.authz == nil {
		return domainmembership.EffectivePermissions{}, shared.ErrForbidden
	}
	return h.authz.ResolveGlobal(ctx, actorID)
}

func (h *Handler) resolveCreateAccess(ctx context.Context, actorID shared.UserID, parentID *string) (domainmembership.EffectivePermissions, error) {
	if parentID == nil {
		return h.resolveGlobalAccess(ctx, actorID)
	}
	targetDivisionID := shared.DivisionID(strings.TrimSpace(*parentID))
	if targetDivisionID == "" {
		return domainmembership.EffectivePermissions{}, shared.ErrEmptyID
	}
	return h.resolveDivisionAccess(ctx, actorID, targetDivisionID)
}

func toTreeNode(divisions []domaindivision.Division) (appdivision.TreeNode, error) {
	if len(divisions) == 0 {
		return appdivision.TreeNode{}, &shared.Error{
			Code:    "not_found.division",
			Message: "division not found",
		}
	}

	nodesByID := make(map[shared.DivisionID]domaindivision.Division, len(divisions))
	childrenByParent := make(map[shared.DivisionID][]shared.DivisionID)
	var rootID shared.DivisionID
	for _, division := range divisions {
		nodesByID[division.ID] = division
		if division.ParentID == nil {
			rootID = division.ID
			continue
		}
		childrenByParent[*division.ParentID] = append(childrenByParent[*division.ParentID], division.ID)
	}

	if rootID == "" {
		return appdivision.TreeNode{}, &shared.Error{
			Code:    "not_found.division",
			Message: "division not found",
		}
	}

	visiting := make(map[shared.DivisionID]bool)
	var build func(shared.DivisionID) (appdivision.TreeNode, error)
	build = func(id shared.DivisionID) (appdivision.TreeNode, error) {
		if visiting[id] {
			return appdivision.TreeNode{}, domaindivision.ErrCycleOnMove
		}
		division, ok := nodesByID[id]
		if !ok {
			return appdivision.TreeNode{}, &shared.Error{
				Code:    "not_found.division",
				Message: "division not found",
			}
		}

		visiting[id] = true
		childIDs := childrenByParent[id]
		children := make([]appdivision.TreeNode, 0, len(childIDs))
		for _, childID := range childIDs {
			childNode, err := build(childID)
			if err != nil {
				return appdivision.TreeNode{}, err
			}
			children = append(children, childNode)
		}
		delete(visiting, id)

		return appdivision.TreeNode{
			Division:       division,
			Children:       children,
			HasChildren:    len(childIDs) > 0,
			ChildrenCount:  len(childIDs),
			PositionsCount: division.PositionsCount,
			MembersCount:   division.MembersCount,
		}, nil
	}

	return build(rootID)
}

func divisionDTO(division domaindivision.Division) map[string]any {
	dto := map[string]any{
		"id":          string(division.ID),
		"short_name":  division.ShortName,
		"full_name":   division.FullName,
		"description": division.Description,
		"is_archived": division.IsArchived,
	}
	if division.RegulationURL != "" {
		dto["regulation_url"] = division.RegulationURL
	}
	if len(division.MediaLinks) > 0 {
		dto["media_links"] = division.MediaLinks
	}
	if division.ParentID != nil {
		dto["parent_id"] = string(*division.ParentID)
	}
	return dto
}

func writeDivisionError(w http.ResponseWriter, r *http.Request, err error) {
	if errors.Is(err, sql.ErrNoRows) {
		writeError(w, r, notFoundDivisionErr())
		return
	}
	if isConflictError(err) {
		writeConflict(w, r, err)
		return
	}
	writeError(w, r, err)
}

func isConflictError(err error) bool {
	return errors.Is(err, domaindivision.ErrRootExists) ||
		errors.Is(err, domaindivision.ErrSelfParent) ||
		errors.Is(err, domaindivision.ErrCycleOnMove) ||
		errors.Is(err, domaindivision.ErrArchivedParent)
}

func writeConflict(w http.ResponseWriter, r *http.Request, err error) {
	payload := map[string]any{
		"error": map[string]any{
			"code":    "conflict.division",
			"message": "conflict",
		},
	}
	var domainErr *shared.Error
	if errors.As(err, &domainErr) {
		payload["error"] = map[string]any{
			"code":    string(domainErr.Code),
			"message": domainErr.Message,
		}
	}
	if requestID := chiMiddleware.GetReqID(r.Context()); requestID != "" {
		payload["request_id"] = requestID
	}
	writeJSON(w, http.StatusConflict, payload)
}

func notFoundDivisionErr() error {
	return &shared.Error{
		Code:    "not_found.division",
		Message: "division not found",
	}
}

var errDivisionArchiveNoop = errors.New("division archive noop")

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

type errorEnvelope struct {
	Error     errorPayload `json:"error"`
	RequestID string       `json:"request_id,omitempty"`
}

type errorPayload struct {
	Code    string         `json:"code"`
	Message string         `json:"message"`
	Details map[string]any `json:"details,omitempty"`
}

func writeError(w http.ResponseWriter, r *http.Request, err error) {
	statusCode := http.StatusInternalServerError
	payload := errorPayload{
		Code:    "internal.error",
		Message: "internal server error",
	}

	var domainErr *shared.Error
	if errors.As(err, &domainErr) {
		statusCode, payload = payloadFromDomainError(*domainErr)
	}

	var fieldErr *shared.FieldError
	if errors.As(err, &fieldErr) {
		statusCode = http.StatusBadRequest
		payload.Code = "validation.field"
		payload.Message = "request validation failed"
		payload.Details = map[string]any{
			"field": fieldErr.Field,
			"kind":  fieldErr.Kind,
		}
		if fieldErr.Limit > 0 {
			payload.Details["limit"] = fieldErr.Limit
		}
	}

	writeErrorEnvelope(w, r, statusCode, payload)
}

func writeErrorEnvelope(w http.ResponseWriter, r *http.Request, statusCode int, payload errorPayload) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	_ = json.NewEncoder(w).Encode(errorEnvelope{
		Error:     payload,
		RequestID: chiMiddleware.GetReqID(r.Context()),
	})
}

func payloadFromDomainError(err shared.Error) (int, errorPayload) {
	status := statusFromDomainCode(err.Code)
	payload := errorPayload{
		Code:    string(err.Code),
		Message: "request cannot be processed",
	}

	if status == http.StatusBadRequest {
		payload.Message = err.Message
		return status, payload
	}
	if status == http.StatusForbidden {
		payload.Message = "operation is forbidden"
		return status, payload
	}
	if status == http.StatusNotFound {
		payload.Message = "resource not found"
		return status, payload
	}
	if status == http.StatusUnauthorized {
		payload.Message = "unauthorized"
		return status, payload
	}
	return status, payload
}

func statusFromDomainCode(code shared.ErrorCode) int {
	switch {
	case strings.HasPrefix(string(code), "access."):
		return http.StatusForbidden
	case strings.HasPrefix(string(code), "auth."):
		return http.StatusUnauthorized
	case strings.HasPrefix(string(code), "not_found."):
		return http.StatusNotFound
	case strings.HasPrefix(string(code), "validation."):
		return http.StatusBadRequest
	case code == "shared.empty_id":
		return http.StatusBadRequest
	default:
		return http.StatusUnprocessableEntity
	}
}
