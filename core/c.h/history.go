package chcore

import (
	"context"
	"errors"
	"strings"
)

var (
	ErrHistoryItemsEmpty             = errors.New("items cannot be empty")
	ErrHistoryItemFieldsRequired     = errors.New("anime_id, name, cover are required")
	ErrSyncHistoryItemFieldsRequired = errors.New("history item fields are required")
)

type HistoryUploadItem struct {
	AnimeID int64
	Name    string
	Cover   string
	AddedAt string
}

type HistoryQueryItem struct {
	ID        int64  `json:"id"`
	AnimeID   int64  `json:"anime_id"`
	Name      string `json:"name"`
	Cover     string `json:"cover"`
	AddedAt   string `json:"added_at"`
	CreatedAt string `json:"created_at"`
}

type RemovedHistoryQueryItem struct {
	AnimeID   int64  `json:"anime_id"`
	Name      string `json:"name"`
	Cover     string `json:"cover"`
	RemovedAt string `json:"removed_at"`
	AddedAt   string `json:"added_at"`
}

type HistoryListResult struct {
	Items        []HistoryQueryItem        `json:"items"`
	RemovedItems []RemovedHistoryQueryItem `json:"removed_items"`
}

func (s *UserService) UploadHistory(ctx context.Context, userID int64, items []HistoryUploadItem) (int, error) {
	if len(items) == 0 {
		return 0, ErrHistoryItemsEmpty
	}

	unique, err := dedupeHistoryItems(items, ErrHistoryItemFieldsRequired)
	if err != nil {
		return 0, err
	}

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback(ctx)

	for _, item := range unique {
		_, err = tx.Exec(ctx, `
			INSERT INTO anime_history_records (user_id, anime_id, anime_name, cover, client_added_at, is_deleted, updated_at)
			VALUES ($1, $2, $3, $4, $5, FALSE, NOW())
			ON CONFLICT (user_id, anime_id) DO UPDATE
			SET anime_name = EXCLUDED.anime_name,
			    cover = EXCLUDED.cover,
			    client_added_at = EXCLUDED.client_added_at,
			    is_deleted = FALSE,
			    updated_at = NOW()
		`, userID, item.AnimeID, item.Name, item.Cover, parseRFC3339Nullable(item.AddedAt))
		if err != nil {
			return 0, err
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return 0, err
	}

	return len(unique), nil
}

func (s *UserService) ListHistory(ctx context.Context, userID int64, limit int) (*HistoryListResult, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT id, anime_id, anime_name, cover,
		       COALESCE(client_added_at::text, '') AS client_added_at,
		       created_at::text
		FROM anime_history_records
		WHERE user_id = $1 AND is_deleted = FALSE
		ORDER BY COALESCE(client_added_at, updated_at, created_at) DESC, updated_at DESC
		LIMIT $2
	`, userID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]HistoryQueryItem, 0, limit)
	for rows.Next() {
		var item HistoryQueryItem
		if err := rows.Scan(&item.ID, &item.AnimeID, &item.Name, &item.Cover, &item.AddedAt, &item.CreatedAt); err != nil {
			return nil, err
		}
		items = append(items, item)
	}

	removedRows, err := s.pool.Query(ctx, `
		SELECT anime_id, anime_name, cover,
		       updated_at::text AS removed_at,
		       COALESCE(client_added_at::text, '') AS added_at
		FROM anime_history_records
		WHERE user_id = $1 AND is_deleted = TRUE
		ORDER BY updated_at DESC
		LIMIT $2
	`, userID, limit)
	if err != nil {
		return nil, err
	}
	defer removedRows.Close()

	removedItems := make([]RemovedHistoryQueryItem, 0, limit)
	for removedRows.Next() {
		var item RemovedHistoryQueryItem
		if err := removedRows.Scan(&item.AnimeID, &item.Name, &item.Cover, &item.RemovedAt, &item.AddedAt); err != nil {
			return nil, err
		}
		removedItems = append(removedItems, item)
	}

	return &HistoryListResult{Items: items, RemovedItems: removedItems}, nil
}

func dedupeHistoryItems(items []HistoryUploadItem, fieldErr error) ([]HistoryUploadItem, error) {
	unique := make([]HistoryUploadItem, 0, len(items))
	seen := make(map[int64]struct{}, len(items))
	for i := len(items) - 1; i >= 0; i-- {
		item := items[i]
		if item.AnimeID == 0 || strings.TrimSpace(item.Name) == "" || strings.TrimSpace(item.Cover) == "" {
			return nil, fieldErr
		}
		if _, ok := seen[item.AnimeID]; ok {
			continue
		}
		seen[item.AnimeID] = struct{}{}
		unique = append(unique, item)
	}

	for i, j := 0, len(unique)-1; i < j; i, j = i+1, j-1 {
		unique[i], unique[j] = unique[j], unique[i]
	}

	return unique, nil
}
