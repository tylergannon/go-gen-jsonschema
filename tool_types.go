//go:build nobuild

package jsonschema

import (
	"context"
	"database/sql"
	"encoding/json"
	"log/slog"
	"time"
)

type Tool interface {
	Name() string
	Description() string
	Parameters() json.RawMessage
	Invoke(ctx context.Context, params string) (string, error)
}

func BuildTool(name, description string, impl any) Tool {
	panic("not implemented")
}

func GetStockPrice(ctx context.Context, symbol string /* the symbol to look up */, date time.Time /* the date to look up */) (string, error) {
	panic("not implemented")
}

var Tool0 = BuildTool(
	"get_stock_price",
	"get the stock price for a given symbol and date",
	GetStockPrice,
)

var Tool1 = BuildTool(
	"find_user",
	"locate user by matching on name or email",
	func(ctx context.Context,
		name string, // regex to match against first and last name
		email string, // match exact for email
		db *sql.DB,
		logger *slog.Logger,
	) (string, error) {
		return "user_id", nil
	},
)

type SomeCoolTool struct {
	SomeCoolField string
	DB            *sql.DB
	Logger        *slog.Logger
}
