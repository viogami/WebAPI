package handlers

import (
	"errors"
	"net/http"

	chcore "webapi/core/c.h"

	"github.com/gin-gonic/gin"
)

type historyUploadItem struct {
	AnimeID int64  `json:"anime_id"`
	Name    string `json:"name"`
	Cover   string `json:"cover"`
	AddedAt string `json:"added_at"`
}

func (h *CHHandler) UploadHistory(c *gin.Context) {
	userID, ok := authedUserID(c)
	if !ok {
		writeError(c, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req struct {
		Items []historyUploadItem `json:"items"`
	}
	if err := decodeJSONBody(c, &req); err != nil {
		writeError(c, http.StatusBadRequest, err.Error())
		return
	}
	items := make([]chcore.HistoryUploadItem, 0, len(req.Items))
	for _, item := range req.Items {
		items = append(items, chcore.HistoryUploadItem{
			AnimeID: item.AnimeID,
			Name:    item.Name,
			Cover:   item.Cover,
			AddedAt: item.AddedAt,
		})
	}

	uploaded, err := h.userService.UploadHistory(c.Request.Context(), userID, items)
	if err != nil {
		switch {
		case errors.Is(err, chcore.ErrHistoryItemsEmpty), errors.Is(err, chcore.ErrHistoryItemFieldsRequired):
			writeError(c, http.StatusBadRequest, err.Error())
		default:
			writeError(c, http.StatusInternalServerError, "insert history failed")
		}
		return
	}

	c.JSON(http.StatusCreated, gin.H{"uploaded": uploaded})
}

func (h *CHHandler) ListHistory(c *gin.Context) {
	userID, ok := authedUserID(c)
	if !ok {
		writeError(c, http.StatusUnauthorized, "unauthorized")
		return
	}

	limit := parseLimit(c.Query("limit"), 50, 1, 200)
	result, err := h.userService.ListHistory(c.Request.Context(), userID, limit)
	if err != nil {
		writeError(c, http.StatusInternalServerError, "query history failed")
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"items":         result.Items,
		"removed_items": result.RemovedItems,
	})
}

func (h *CHHandler) Sync(c *gin.Context) {
	userID, ok := authedUserID(c)
	if !ok {
		writeError(c, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req struct {
		History []historyUploadItem `json:"history"`
		Rank    *rankUploadRequest  `json:"rank"`
	}
	if err := decodeJSONBody(c, &req); err != nil {
		writeError(c, http.StatusBadRequest, err.Error())
		return
	}
	historyItems := make([]chcore.HistoryUploadItem, 0, len(req.History))
	for _, item := range req.History {
		historyItems = append(historyItems, chcore.HistoryUploadItem{
			AnimeID: item.AnimeID,
			Name:    item.Name,
			Cover:   item.Cover,
			AddedAt: item.AddedAt,
		})
	}

	var rankInput *chcore.RankInput
	if req.Rank != nil {
		rankInput = &chcore.RankInput{
			Title:         req.Rank.Title,
			TierBoardName: req.Rank.TierBoardName,
			GridBoardName: req.Rank.GridBoardName,
			Payload:       req.Rank.Payload,
		}
	}

	result, err := h.rankService.Sync(c.Request.Context(), userID, chcore.SyncInput{
		History: historyItems,
		Rank:    rankInput,
	})
	if err != nil {
		switch {
		case errors.Is(err, chcore.ErrHistoryOrRank), errors.Is(err, chcore.ErrSyncHistoryItemFieldsRequired), errors.Is(err, chcore.ErrInvalidRankPayload):
			writeError(c, http.StatusBadRequest, err.Error())
		default:
			writeError(c, http.StatusInternalServerError, "commit tx failed")
		}
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"history_uploaded": result.HistoryUploaded,
		"rank_id":          result.RankID,
	})
}
