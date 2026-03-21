package chcore

import (
	"context"
	"encoding/json"
	"errors"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrRankFieldsRequired = errors.New("title, tier_board_name, grid_board_name are required")
	ErrPayloadRequired    = errors.New("payload is required")
	ErrHistoryOrRank      = errors.New("history or rank is required")
	ErrInvalidRankPayload = errors.New("invalid rank payload")
)

type RankService struct {
	pool *pgxpool.Pool
}

type RankInput struct {
	Title         string
	TierBoardName string
	GridBoardName string
	Payload       json.RawMessage
}

type RankItem struct {
	ID            int64           `json:"id"`
	Title         string          `json:"title"`
	TierBoardName string          `json:"tier_board_name"`
	GridBoardName string          `json:"grid_board_name"`
	Payload       json.RawMessage `json:"payload"`
	CreatedAt     string          `json:"created_at"`
}

type SyncInput struct {
	History []HistoryUploadItem
	Rank    *RankInput
}

type SyncResult struct {
	HistoryUploaded int
	RankID          *int64
}

func NewRankService(pool *pgxpool.Pool) *RankService {
	return &RankService{pool: pool}
}

func (s *RankService) CreateRank(ctx context.Context, userID int64, input RankInput) (int64, error) {
	input.Title = strings.TrimSpace(input.Title)
	input.TierBoardName = strings.TrimSpace(input.TierBoardName)
	input.GridBoardName = strings.TrimSpace(input.GridBoardName)
	if input.Title == "" || input.TierBoardName == "" || input.GridBoardName == "" {
		return 0, ErrRankFieldsRequired
	}
	if len(input.Payload) == 0 {
		return 0, ErrPayloadRequired
	}

	var id int64
	err := s.pool.QueryRow(ctx, `
		INSERT INTO rank_snapshots (user_id, title, tier_board_name, grid_board_name, payload)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id
	`, userID, input.Title, input.TierBoardName, input.GridBoardName, input.Payload).Scan(&id)
	if err != nil {
		return 0, err
	}

	return id, nil
}

func (s *RankService) ListRank(ctx context.Context, userID int64, limit int) ([]RankItem, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT id, title, tier_board_name, grid_board_name, payload, created_at::text
		FROM rank_snapshots
		WHERE user_id = $1
		ORDER BY created_at DESC
		LIMIT $2
	`, userID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]RankItem, 0, limit)
	for rows.Next() {
		var item RankItem
		if err := rows.Scan(&item.ID, &item.Title, &item.TierBoardName, &item.GridBoardName, &item.Payload, &item.CreatedAt); err != nil {
			return nil, err
		}
		items = append(items, item)
	}

	return items, nil
}

func (s *RankService) LatestRank(ctx context.Context, userID int64) (*RankItem, error) {
	var item RankItem
	err := s.pool.QueryRow(ctx, `
		SELECT id, title, tier_board_name, grid_board_name, payload, created_at::text
		FROM rank_snapshots
		WHERE user_id = $1
		ORDER BY created_at DESC
		LIMIT 1
	`, userID).Scan(&item.ID, &item.Title, &item.TierBoardName, &item.GridBoardName, &item.Payload, &item.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	return &item, nil
}

func (s *RankService) Sync(ctx context.Context, userID int64, input SyncInput) (*SyncResult, error) {
	if len(input.History) == 0 && input.Rank == nil {
		return nil, ErrHistoryOrRank
	}

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	historyCount := 0
	for _, item := range input.History {
		if item.AnimeID == 0 || strings.TrimSpace(item.Name) == "" || strings.TrimSpace(item.Cover) == "" {
			return nil, ErrSyncHistoryItemFieldsRequired
		}
		_, err = tx.Exec(ctx, `
			INSERT INTO anime_history_records (user_id, anime_id, anime_name, cover, client_added_at)
			VALUES ($1, $2, $3, $4, $5)
		`, userID, item.AnimeID, item.Name, item.Cover, parseRFC3339Nullable(item.AddedAt))
		if err != nil {
			return nil, err
		}
		historyCount++
	}

	var rankID *int64
	if input.Rank != nil {
		rank := RankInput{
			Title:         strings.TrimSpace(input.Rank.Title),
			TierBoardName: strings.TrimSpace(input.Rank.TierBoardName),
			GridBoardName: strings.TrimSpace(input.Rank.GridBoardName),
			Payload:       input.Rank.Payload,
		}
		if rank.Title == "" || rank.TierBoardName == "" || rank.GridBoardName == "" || len(rank.Payload) == 0 {
			return nil, ErrInvalidRankPayload
		}

		var id int64
		err = tx.QueryRow(ctx, `
			INSERT INTO rank_snapshots (user_id, title, tier_board_name, grid_board_name, payload)
			VALUES ($1, $2, $3, $4, $5)
			RETURNING id
		`, userID, rank.Title, rank.TierBoardName, rank.GridBoardName, rank.Payload).Scan(&id)
		if err != nil {
			return nil, err
		}
		rankID = &id
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	return &SyncResult{HistoryUploaded: historyCount, RankID: rankID}, nil
}
