package moex

import (
	"context"
	"fmt"
	"net/url"
)

// Listed сообщает, есть ли тикер в каталоге акций TQBR.
// Пустой список — бумага не найдена. Сбой ISS возвращается как ошибка.
func (c *Client) Listed(ctx context.Context, ticker string) (bool, error) {
	if err := ident("security", ticker); err != nil {
		return false, err
	}
	path := "/iss/engines/stock/markets/shares/boards/TQBR/securities/" + ticker + ".json"
	query := url.Values{}
	query.Set("iss.meta", "off")
	query.Set("iss.only", "securities,securities.cursor")
	query.Set("securities.columns", "SECID,SHORTNAME")

	pages, err := c.collect(ctx, path, query, "securities", "SECID")
	if err != nil {
		return false, fmt.Errorf("moex catalog: %w", err)
	}
	rows, err := parseSecurities(pages)
	if err != nil {
		return false, fmt.Errorf("moex catalog: %w", err)
	}
	for _, row := range rows {
		if row.SecID == ticker {
			return true, nil
		}
	}
	return false, nil
}
