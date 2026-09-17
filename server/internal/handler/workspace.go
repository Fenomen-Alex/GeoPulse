package handler

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/alex/geopulse/server/internal/auth"
	"github.com/alex/geopulse/server/internal/db"
)

// WorkspaceStore is the persistence dependency for the workspace endpoints. The
// concrete *db.DB satisfies it; tests can supply a stub.
type WorkspaceStore interface {
	EnsureUser(userID string) error
	SaveWorkspace(w db.Workspace) error
	ListWorkspaces(userID string) ([]db.Workspace, error)
}

// WorkspaceHandler persists client workbench state (origin, tool, travel mode,
// route) so a user's inputs survive page reloads and new sessions.
type WorkspaceHandler struct {
	store WorkspaceStore
}

// NewWorkspaceHandler builds the handler. A nil store degrades the endpoint to
// 503 so the client can silently skip persistence when no DB is configured.
func NewWorkspaceHandler(store WorkspaceStore) *WorkspaceHandler {
	return &WorkspaceHandler{store: store}
}

type workspaceResponse struct {
	ID         string            `json:"id"`
	Title      string            `json:"title"`
	CenterLat  float64           `json:"center_lat"`
	CenterLng  float64           `json:"center_lng"`
	Zoom       float64           `json:"zoom"`
	LayersJSON json.RawMessage   `json:"layers_json,omitempty"`
	CreatedAt  string            `json:"created_at,omitempty"`
}

// List returns the authenticated user's saved workspaces.
func (h *WorkspaceHandler) List(w http.ResponseWriter, r *http.Request) {
	if h.store == nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]string{"error": "persistence unavailable"})
		return
	}

	userID, _ := auth.GetUserIDFromContext(r.Context())
	workspaces, err := h.store.ListWorkspaces(userID)
	if err != nil {
		log.Printf("workspaces: list failed: %v", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to load workspaces"})
		return
	}

	results := make([]workspaceResponse, 0, len(workspaces))
	for _, w := range workspaces {
		results = append(results, workspaceResponse{
			ID:         w.ID,
			Title:      w.Title,
			CenterLat:  w.CenterLat,
			CenterLng:  w.CenterLng,
			Zoom:       w.Zoom,
			LayersJSON: json.RawMessage(w.LayersJSON),
			CreatedAt:  w.CreatedAt,
		})
	}
	writeJSON(w, http.StatusOK, map[string]any{"workspaces": results})
}

// Save upserts a workbench snapshot for the authenticated user. The id field is
// optional: clients may pass an existing id to overwrite it, or omit it to have
// one generated.
func (h *WorkspaceHandler) Save(w http.ResponseWriter, r *http.Request) {
	if h.store == nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]string{"error": "persistence unavailable"})
		return
	}

	userID, _ := auth.GetUserIDFromContext(r.Context())

	var body struct {
		ID         string          `json:"id"`
		Title      string          `json:"title"`
		CenterLat  float64         `json:"center_lat"`
		CenterLng  float64         `json:"center_lng"`
		Zoom       float64         `json:"zoom"`
		LayersJSON json.RawMessage `json:"layers_json"`
	}
	decoder := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&body); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	if body.ID == "" {
		body.ID = newID()
	}
	if body.Title == "" {
		body.Title = "Untitled workspace"
	}
	layers := body.LayersJSON
	if len(layers) == 0 {
		layers = json.RawMessage("{}")
	}

	// The workspaces FK requires a users row; seed it idempotently.
	if err := h.store.EnsureUser(userID); err != nil {
		log.Printf("workspaces: EnsureUser failed: %v", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to persist workspace"})
		return
	}

	workspace := db.Workspace{
		ID:         body.ID,
		UserID:     userID,
		Title:      body.Title,
		CenterLat:  body.CenterLat,
		CenterLng:  body.CenterLng,
		Zoom:       body.Zoom,
		LayersJSON: string(layers),
	}
	if err := h.store.SaveWorkspace(workspace); err != nil {
		log.Printf("workspaces: save failed: %v", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to persist workspace"})
		return
	}

	writeJSON(w, http.StatusOK, workspaceResponse{
		ID:         workspace.ID,
		Title:      workspace.Title,
		CenterLat:  workspace.CenterLat,
		CenterLng:  workspace.CenterLng,
		Zoom:       workspace.Zoom,
		LayersJSON: layers,
	})
}