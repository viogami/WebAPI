package server

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/jackc/pgx/v5"
)

type rankUploadRequest struct {
	Title         string          `json:"title"`
	TierBoardName string          `json:"tier_board_name"`
	GridBoardName string          `json:"grid_board_name"`
	Payload       json.RawMessage `json:"payload"`
}

func (s *Server) handleRank(w http.ResponseWriter, r *http.Request, userID int64) {
	switch r.Method {
	case http.MethodPost:
		s.createRank(w, r, userID)
	case http.MethodGet:
		s.listRank(w, r, userID)
	default:
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

func (s *Server) createRank(w http.ResponseWriter, r *http.Request, userID int64) {
	var req rankUploadRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	req.Title = strings.TrimSpace(req.Title)
	req.TierBoardName = strings.TrimSpace(req.TierBoardName)
	req.GridBoardName = strings.TrimSpace(req.GridBoardName)

	if req.Title == "" || req.TierBoardName == "" || req.GridBoardName == "" {
		writeError(w, http.StatusBadRequest, "title, tier_board_name, grid_board_name are required")
		return
	}
	if len(req.Payload) == 0 {
		writeError(w, http.StatusBadRequest, "payload is required")
		return
	}

	var id int64
	err := s.pool.QueryRow(r.Context(), `
		INSERT INTO rank_snapshots (user_id, title, tier_board_name, grid_board_name, payload)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id
	`, userID, req.Title, req.TierBoardName, req.GridBoardName, req.Payload).Scan(&id)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "insert rank failed")
		return
	}

	writeJSON(w, http.StatusCreated, map[string]any{"id": id})
}

func (s *Server) listRank(w http.ResponseWriter, r *http.Request, userID int64) {
	limit := parseLimit(r.URL.Query().Get("limit"), 20, 1, 100)

	rows, err := s.pool.Query(r.Context(), `
		SELECT id, title, tier_board_name, grid_board_name, payload, created_at::text
		FROM rank_snapshots
		WHERE user_id = $1
		ORDER BY created_at DESC
		LIMIT $2
	`, userID, limit)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "query rank failed")
		return
	}
	defer rows.Close()

	type rankItem struct {
		ID            int64           `json:"id"`
		Title         string          `json:"title"`
		TierBoardName string          `json:"tier_board_name"`
		GridBoardName string          `json:"grid_board_name"`
		Payload       json.RawMessage `json:"payload"`
		CreatedAt     string          `json:"created_at"`
	}
	items := make([]rankItem, 0, limit)
	for rows.Next() {
		var item rankItem
		if err := rows.Scan(&item.ID, &item.Title, &item.TierBoardName, &item.GridBoardName, &item.Payload, &item.CreatedAt); err != nil {
			writeError(w, http.StatusInternalServerError, "scan rank failed")
			return
		}
		items = append(items, item)
	}

	writeJSON(w, http.StatusOK, map[string]any{"items": items})
}

func (s *Server) handleLatestRank(w http.ResponseWriter, r *http.Request, userID int64) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	var id int64
	var title, tierBoardName, gridBoardName string
	var payload json.RawMessage
	var createdAt string
	err := s.pool.QueryRow(r.Context(), `
		SELECT id, title, tier_board_name, grid_board_name, payload, created_at::text
		FROM rank_snapshots
		WHERE user_id = $1
		ORDER BY created_at DESC
		LIMIT 1
	`, userID).Scan(&id, &title, &tierBoardName, &gridBoardName, &payload, &createdAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			writeJSON(w, http.StatusOK, map[string]any{"item": nil})
			return
		}
		writeError(w, http.StatusInternalServerError, "query latest rank failed")
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"item": map[string]any{
			"id":              id,
			"title":           title,
			"tier_board_name": tierBoardName,
			"grid_board_name": gridBoardName,
			"payload":         payload,
			"created_at":      createdAt,
		},
	})
}
