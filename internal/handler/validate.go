package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"

	"github.com/MaksimKorneplod/MyInvestor/internal/repository/moex"
)

const maxBodyBytes = 1 << 16 // 64 KiB — с запасом на 200 тикеров

var tickerRe = regexp.MustCompile(`^[A-Za-z0-9._-]{1,32}$`)

// allowedMarkets — whitelist engine/market/board.
// Всё, чего здесь нет, отбиваем 400 до похода в ISS.
var allowedMarkets = map[string]map[string]map[string]struct{}{
	"stock": {
		"shares": {"": {}, "TQBR": {}, "TQTF": {}},
		"bonds":  {"": {}, "TQCB": {}, "TQOB": {}},
	},
	"currency": {
		"selt": {"": {}},
	},
}

// validateInstrument проверяет engine/market/board по whitelist.
// Возвращает доменный moex.Instrument без Security — Security
// проставляет вызывающий handler.
func validateInstrument(engine, market, board string) (moex.Instrument, error) {
	byMarket, ok := allowedMarkets[engine]
	if !ok {
		return moex.Instrument{}, fmt.Errorf("engine %q not allowed", engine)
	}
	boards, ok := byMarket[market]
	if !ok {
		return moex.Instrument{}, fmt.Errorf("market %q not allowed", market)
	}
	if _, ok := boards[board]; !ok {
		return moex.Instrument{}, fmt.Errorf("board %q not allowed", board)
	}
	return moex.Instrument{Engine: engine, Market: market, Board: board}, nil
}

func validateTicker(t string) error {
	if !tickerRe.MatchString(t) {
		return fmt.Errorf("bad ticker %q", t)
	}
	return nil
}

// decodeBatch — строгий парсер тела POST /v1/quotes:batch:
//   - ограничение размера (MaxBytesReader)
//   - DisallowUnknownFields
//   - ровно одно JSON-значение, никакого trailing data
func decodeBatch(w http.ResponseWriter, r *http.Request) (BatchRequest, error) {
	r.Body = http.MaxBytesReader(w, r.Body, maxBodyBytes)
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()

	var req BatchRequest
	if err := dec.Decode(&req); err != nil {
		return req, err
	}
	// Проверяем, что после первого значения — только EOF.
	if err := dec.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		return req, errors.New("trailing data after body")
	}
	return req, nil
}

// normalizeBoard — «пустой» board сводим к одному виду.
// "" и " " — одно и то же; это важно для кэш-ключа в service.
func normalizeBoard(b string) string {
	return strings.TrimSpace(b)
}
