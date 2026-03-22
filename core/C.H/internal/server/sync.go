package server

import (
	"net/http"
	"strings"
)

func (s *Server) handleSync(w http.ResponseWriter, r *http.Request, userID int64) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	var req struct {
		History        []historyUploadItem `json:"history"`
		RemovedHistory []struct {
			AnimeID int64 `json:"anime_id"`
		} `json:"removed_history"`
		Rank *rankUploadRequest `json:"rank"`
	}
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	if len(req.History) == 0 && len(req.RemovedHistory) == 0 && req.Rank == nil {
		writeError(w, http.StatusBadRequest, "history or rank is required")
		return
	}

	tx, err := s.pool.Begin(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "begin tx failed")
		return
	}
	defer tx.Rollback(r.Context())

	historyAddedOrRestored := 0
	historyRemoved := 0
	unique := make([]historyUploadItem, 0, len(req.History))
	seen := make(map[int64]struct{}, len(req.History))
	for i := len(req.History) - 1; i >= 0; i-- {
		item := req.History[i]
		if item.AnimeID == 0 || strings.TrimSpace(item.Name) == "" {
			writeError(w, http.StatusBadRequest, "history item fields are required")
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
		historyAddedOrRestored += int(updateTag.RowsAffected())

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
		historyAddedOrRestored += int(insertTag.RowsAffected())
	}

	removedIDs := make([]int64, 0, len(req.RemovedHistory))
	removedSeen := make(map[int64]struct{}, len(req.RemovedHistory))
	for _, item := range req.RemovedHistory {
		if item.AnimeID == 0 {
			continue
		}
		if _, active := seen[item.AnimeID]; active {
			continue
		}
		if _, ok := removedSeen[item.AnimeID]; ok {
			continue
		}
		removedSeen[item.AnimeID] = struct{}{}
		removedIDs = append(removedIDs, item.AnimeID)
	}

	if len(removedIDs) > 0 {
		deleteTag, err := tx.Exec(r.Context(), `
			UPDATE anime_history_records
			SET is_deleted = TRUE,
			    updated_at = NOW()
			WHERE user_id = $1
			  AND is_deleted = FALSE
			  AND anime_id = ANY($2::BIGINT[])
		`, userID, removedIDs)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "mark deleted failed")
			return
		}
		historyRemoved += int(deleteTag.RowsAffected())
	}

	var rankID *int64
	if req.Rank != nil {
		req.Rank.Title = strings.TrimSpace(req.Rank.Title)
		req.Rank.TierBoardName = strings.TrimSpace(req.Rank.TierBoardName)
		req.Rank.GridBoardName = strings.TrimSpace(req.Rank.GridBoardName)
		if req.Rank.Title == "" || req.Rank.TierBoardName == "" || req.Rank.GridBoardName == "" || len(req.Rank.Payload) == 0 {
			writeError(w, http.StatusBadRequest, "invalid rank payload")
			return
		}
		var id int64
		err := tx.QueryRow(r.Context(), `
			INSERT INTO rank_snapshots (user_id, title, tier_board_name, grid_board_name, payload)
			VALUES ($1, $2, $3, $4, $5)
			RETURNING id
		`, userID, req.Rank.Title, req.Rank.TierBoardName, req.Rank.GridBoardName, req.Rank.Payload).Scan(&id)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "insert rank failed")
			return
		}
		rankID = &id
	}

	if err := tx.Commit(r.Context()); err != nil {
		writeError(w, http.StatusInternalServerError, "commit tx failed")
		return
	}

	writeJSON(w, http.StatusCreated, map[string]any{
		"history_uploaded":          historyAddedOrRestored,
		"history_removed":           historyRemoved,
		"history_changed_total":     historyAddedOrRestored + historyRemoved,
		"rank_id":                   rankID,
	})
}
