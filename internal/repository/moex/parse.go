package moex

import (
	"encoding/json"
	"fmt"
	"regexp"
	"sort"
	"time"
)

const (
	dateLayout = "2006-01-02"
	timeLayout = "2006-01-02 15:04:05"
	maxPages   = 100
)

var token = regexp.MustCompile(`^[A-Za-z0-9._-]{1,32}$`)

func ident(kind, value string) error {
	if !token.MatchString(value) {
		return fmt.Errorf("%w: %s %q", ErrRequest, kind, value)
	}
	return nil
}

// table — один блок ISS: columns и data в одном порядке.
type table struct {
	Columns []string            `json:"columns"`
	Data    [][]json.RawMessage `json:"data"`
}

func decodeTables(body []byte) (map[string]table, error) {
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(body, &raw); err != nil || raw == nil {
		return nil, &ParseError{Reason: "json"}
	}
	out := make(map[string]table, len(raw))
	for name, payload := range raw {
		if len(payload) == 0 || payload[0] != '{' {
			continue
		}
		var block table
		if err := json.Unmarshal(payload, &block); err != nil {
			return nil, &ParseError{Reason: name}
		}
		out[name] = block
	}
	return out, nil
}

func (t table) indexes(names ...string) (map[string]int, error) {
	index := make(map[string]int, len(names))
	for _, name := range names {
		found := false
		for i, column := range t.Columns {
			if column == name {
				index[name] = i
				found = true
				break
			}
		}
		if !found {
			return nil, &ParseError{Reason: "missing column " + name}
		}
	}
	return index, nil
}

func at(row []json.RawMessage, index int) (json.RawMessage, error) {
	if index < 0 || index >= len(row) {
		return nil, &ParseError{Reason: "short row"}
	}
	return row[index], nil
}

func isNull(raw json.RawMessage) bool {
	return len(raw) == 0 || string(raw) == "null"
}

func cellString(raw json.RawMessage) (string, bool, error) {
	if isNull(raw) {
		return "", false, nil
	}
	var text string
	if err := json.Unmarshal(raw, &text); err != nil {
		return "", false, &ParseError{Reason: "string cell"}
	}
	return text, true, nil
}

func cellDecimal(raw json.RawMessage) (Decimal, bool, error) {
	if isNull(raw) {
		return Decimal{}, false, nil
	}
	var number json.Number
	if err := json.Unmarshal(raw, &number); err != nil || number == "" {
		return Decimal{}, false, &ParseError{Reason: "number cell"}
	}
	if _, err := number.Float64(); err != nil {
		return Decimal{}, false, &ParseError{Reason: "number cell"}
	}
	return Decimal{text: number.String()}, true, nil
}

func cellInt(raw json.RawMessage) (int, error) {
	number, ok, err := cellDecimal(raw)
	if err != nil || !ok {
		return 0, &ParseError{Reason: "cursor"}
	}
	value, err := strconvAtoi(number.text)
	if err != nil {
		return 0, &ParseError{Reason: "cursor"}
	}
	return value, nil
}

func cursorNext(block table, current int) (int, bool, error) {
	if len(block.Data) == 0 {
		return 0, false, nil
	}
	index, err := block.indexes("INDEX", "TOTAL", "PAGESIZE")
	if err != nil {
		return 0, false, err
	}
	row := block.Data[0]
	pos, err := requiredInt(row, index["INDEX"])
	if err != nil {
		return 0, false, err
	}
	total, err := requiredInt(row, index["TOTAL"])
	if err != nil {
		return 0, false, err
	}
	pageSize, err := requiredInt(row, index["PAGESIZE"])
	if err != nil {
		return 0, false, err
	}
	if pageSize <= 0 {
		return 0, false, nil
	}
	next := pos + pageSize
	if next <= current || next >= total {
		return 0, false, nil
	}
	return next, true, nil
}

func requiredInt(row []json.RawMessage, index int) (int, error) {
	raw, err := at(row, index)
	if err != nil {
		return 0, err
	}
	return cellInt(raw)
}

func parseSecurities(pages []table) ([]Security, error) {
	out := make([]Security, 0)
	for _, page := range pages {
		index, err := page.indexes("SECID", "SHORTNAME")
		if err != nil {
			return nil, err
		}
		for _, row := range page.Data {
			rawID, err := at(row, index["SECID"])
			if err != nil {
				return nil, err
			}
			secid, ok, err := cellString(rawID)
			if err != nil {
				return nil, err
			}
			if !ok || secid == "" {
				return nil, &ParseError{Reason: "SECID"}
			}
			rawName, err := at(row, index["SHORTNAME"])
			if err != nil {
				return nil, err
			}
			name, ok, err := cellString(rawName)
			if err != nil {
				return nil, err
			}
			if !ok || name == "" {
				name = secid
			}
			out = append(out, Security{SecID: secid, ShortName: name})
		}
	}
	return out, nil
}

func parseQuotes(pages []table) ([]Quote, error) {
	out := make([]Quote, 0)
	for _, page := range pages {
		index, err := page.indexes("SECID", "LAST", "LASTCHANGEPRCNT")
		if err != nil {
			return nil, err
		}
		for _, row := range page.Data {
			rawID, err := at(row, index["SECID"])
			if err != nil {
				return nil, err
			}
			secid, ok, err := cellString(rawID)
			if err != nil {
				return nil, err
			}
			if !ok || secid == "" {
				return nil, &ParseError{Reason: "SECID"}
			}
			rawLast, err := at(row, index["LAST"])
			if err != nil {
				return nil, err
			}
			last, hasLast, err := cellDecimal(rawLast)
			if err != nil {
				return nil, err
			}
			rawChange, err := at(row, index["LASTCHANGEPRCNT"])
			if err != nil {
				return nil, err
			}
			change, hasChange, err := cellDecimal(rawChange)
			if err != nil {
				return nil, err
			}
			if !hasLast {
				continue
			}
			price := last
			quote := Quote{SecID: secid, Last: &price}
			if hasChange {
				delta := change
				quote.LastChangePercent = &delta
			}
			out = append(out, quote)
		}
	}
	return out, nil
}

func parseHistory(pages []table) ([]HistoryRow, error) {
	out := make([]HistoryRow, 0)
	for _, page := range pages {
		index, err := page.indexes("begin", "open", "close")
		if err != nil {
			return nil, err
		}
		for _, row := range page.Data {
			rawBegin, err := at(row, index["begin"])
			if err != nil {
				return nil, err
			}
			beginText, ok, err := cellString(rawBegin)
			if err != nil {
				return nil, err
			}
			if !ok {
				return nil, &ParseError{Reason: "begin"}
			}
			begin, err := time.ParseInLocation(timeLayout, beginText, moscow)
			if err != nil {
				return nil, &ParseError{Reason: "begin"}
			}
			rawOpen, err := at(row, index["open"])
			if err != nil {
				return nil, err
			}
			open, ok, err := cellDecimal(rawOpen)
			if err != nil {
				return nil, err
			}
			if !ok {
				return nil, &ParseError{Reason: "open"}
			}
			rawClose, err := at(row, index["close"])
			if err != nil {
				return nil, err
			}
			closePrice, ok, err := cellDecimal(rawClose)
			if err != nil {
				return nil, err
			}
			if !ok {
				return nil, &ParseError{Reason: "close"}
			}
			out = append(out, HistoryRow{Begin: begin, Open: open, Close: closePrice})
		}
	}
	sort.Slice(out, func(i, j int) bool {
		return out[i].Begin.Before(out[j].Begin)
	})
	return out, nil
}

func strconvAtoi(text string) (int, error) {
	var number json.Number = json.Number(text)
	value, err := number.Float64()
	if err != nil {
		return 0, err
	}
	return int(value), nil
}
