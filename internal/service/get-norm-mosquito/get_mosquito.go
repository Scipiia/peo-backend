package get_norm_mosquito

import (
	"context"
	"fmt"
	"log/slog"
	"time"
	"vue-golang/internal/storage"
)

type MosquitoStorage interface {
	UpsertMoskitOrder(ctx context.Context, order storage.MosquitoOrderDTO, externalID string) (int64, error)
	MoscuiteDem(ctx context.Context, since time.Time) ([]storage.MosquitoOrderDTO, error)
}

type MosquitoService struct {
	storage MosquitoStorage
}

func NewMosquitoService(storage MosquitoStorage) *MosquitoService {
	return &MosquitoService{storage: storage}
}

func (s *MosquitoService) Sync(ctx context.Context, since time.Time) error {
	orders, err := s.storage.MoscuiteDem(ctx, since)
	if err != nil {
		return fmt.Errorf("ошибка получения москиток в сервисе %w", err)
	}

	//slog.Info("orders count ", len(orders))
	//externalID := "moskit_КП-123_4"

	for _, order := range orders {
		externalID := fmt.Sprintf("moskit_%s_%d", order.OrderNum, order.ClassID)

		// Опционально: лог для отладки
		slog.Debug("syncing mosquito order",
			"order_num", order.OrderNum,
			"class_id", order.ClassID,
			"external_id", externalID)

		_, err := s.storage.UpsertMoskitOrder(ctx, order, externalID)
		if err != nil {
			slog.Error("failed to upsert mosquito order",
				"order_num", order.OrderNum,
				"external_id", externalID,
				"err", err)
			return fmt.Errorf("ошибка вставки новых москиток в dem_product_instance в сервисе %w", err)
		}
	}

	return nil
}
