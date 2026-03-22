package server

import (
	"net/http"
	"strings"
)

type removedHistoryQueryItem struct {
	AnimeID   int64  `json:"anime_id"`
	Name      string `json:"name"`
	Cover     string `json:"cover"`
	RemovedAt string `json:"removed_at"`
	AddedAt   string `json:"added_at"`
}

func (s *Server) handleHistory(w http.ResponseWriter, r *http.Request, userID int64) {
	switch r.Method {
	case http.MethodPost:
		s.uploadHistory(w, r, userID)
	case http.MethodGet:
		s.listHistory(w, r, userID)
	default:
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

func (s *Server) uploadHistory(w http.ResponseWriter, r *http.Request, userID int64) {
	var req struct {
		Items []historyUploadItem `json:"items"`
	}
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	if len(req.Items) == 0 {
		writeError(w, http.StatusBadRequest, "items cannot be empty")
		return
	}

	tx, err := s.pool.Begin(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "begin tx failed")
		return
	}
	defer tx.Rollback(r.Context())

	unique := make([]historyUploadItem, 0, len(req.Items))
	seen := make(map[int64]struct{}, len(req.Items))
	for i := len(req.Items) - 1; i >= 0; i-- {
		item := req.Items[i]
		if item.AnimeID == 0 || strings.TrimSpace(item.Name) == "" {
			writeError(w, http.StatusBadRequest, "anime_id and name are required")
			return
		}
		item.Name = strings.TrimSpace(item.Name)
		item.Cover = strings.TrimSpace(item.Cover)
		if _, ok := seen[item.AnimeID]; ok {
			continue
		}
		seen[item.AnimeID] = struct{}{}
		unique = append(unique, item)
	}

	for i, j := 0, len(unique)-1; i < j; i, j = i+1, j-1 {
		unique[i], unique[j] = unique[j], unique[i]
	}

	writtenCount := 0
	for _, item := range unique {
		clientAddedAt := parseRFC3339Nullable(item.AddedAt)

		updateTag, err := tx.Exec(r.Context(), `
			UPDATE anime_history_records
			SET anime_name = $3,
			    cover = $4,
			    client_added_at = $5,
			    is_deleted = FALSE,
			    updated_at = NOW()
			WHERE user_id = $1
			  AND anime_id = $2
			  AND (
			    is_deleted = TRUE
			    OR anime_name IS DISTINCT FROM $3
			    OR cover IS DISTINCT FROM $4
			    OR client_added_at IS DISTINCT FROM $5
			  )
		`, userID, item.AnimeID, item.Name, item.Cover, clientAddedAt)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "update history failed")
			return
		}
		writtenCount += int(updateTag.RowsAffected())

		if updateTag.RowsAffected() > 0 {
			continue
		}

		insertTag, err := tx.Exec(r.Context(), `
			INSERT INTO anime_history_records (user_id, anime_id, anime_name, cover, client_added_at, is_deleted, updated_at)
			SELECT $1, $2, $3, $4, $5, FALSE, NOW()
			WHERE NOT EXISTS (
				SELECT 1
				FROM anime_history_records
				WHERE user_id = $1 AND anime_id = $2
			)
		`, userID, item.AnimeID, item.Name, item.Cover, clientAddedAt)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "insert history failed")
			return
		}
		writtenCount += int(insertTag.RowsAffected())
	}

	if err := tx.Commit(r.Context()); err != nil {
		writeError(w, http.StatusInternalServerError, "commit failed")
		return
	}

	writeJSON(w, http.StatusCreated, map[string]any{"uploaded": writtenCount})
}

func (s *Server) listHistory(w http.ResponseWriter, r *http.Request, userID int64) {
	limit := parseLimit(r.URL.Query().Get("limit"), 50, 1, 200)

	rows, err := s.pool.Query(r.Context(), `
		WITH ranked AS (
			SELECT id, anime_id, anime_name, cover,
			       COALESCE(client_added_at::text, '') AS client_added_at,
			       created_at::text AS created_at,
			       updated_at::text AS updated_at,
			       is_deleted,
			       ROW_NUMBER() OVER (
					PARTITION BY is_deleted
					ORDER BY COALESCE(client_added_at, updated_at, created_at) DESC, updated_at DESC
			   ) AS rn
			FROM anime_history_records
			WHERE user_id = $1
		)
		SELECT id, anime_id, anime_name, cover, client_added_at, created_at, updated_at, is_deleted
		FROM ranked
		WHERE rn <= $2
		ORDER BY is_deleted ASC, rn ASC
	`, userID, limit)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "query history failed")
		return
	}
	defer rows.Close()

	items := make([]historyQueryItem, 0, limit)
	removedItems := make([]removedHistoryQueryItem, 0, limit)
	for rows.Next() {
		var item historyQueryItem
		var removed removedHistoryQueryItem
		var updatedAt string
		var isDeleted bool
		if err := rows.Scan(&item.ID, &item.AnimeID, &item.Name, &item.Cover, &item.AddedAt, &item.CreatedAt, &updatedAt, &isDeleted); err != nil {
			writeError(w, http.StatusInternalServerError, "scan history failed")
			return
		}

		if isDeleted {
			removed.AnimeID = item.AnimeID
			removed.Name = item.Name
			removed.Cover = item.Cover
			removed.RemovedAt = updatedAt
			removed.AddedAt = item.AddedAt
			removedItems = append(removedItems, removed)
			continue
		}

		items = append(items, item)
	}

	if err := rows.Err(); err != nil {
		writeError(w, http.StatusInternalServerError, "iterate history failed")
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"items":         items,
		"removed_items": removedItems,
	})
}
