package _interface

import (
	"context"

	"github.com/CodeMachine0121/trading-board/backend/internal/domain/models/entities"
)

//go:generate go tool mockgen -source=i_order_fill_repository.go -destination=mocks/mock_i_order_fill_repository.go -package=mocks

// IOrderFillRepository appends fills; v1 never edits or deletes one.
type IOrderFillRepository interface {
	Create(executionContext context.Context, fill entities.OrderFill) (entities.OrderFill, error)
}
