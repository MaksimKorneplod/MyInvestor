package moex

import (
	"context"
	"io"
	"net/http"
	"testing"
)

func TestChartOHLC(t *testing.T) {
	t.Parallel()
	client := newClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("start") != "0" {
			_, _ = io.WriteString(w, `{"candles":{"columns":["begin","open","high","low","close"],"data":[]}}`)
			return
		}
		_, _ = io.WriteString(w, `{"candles":{"columns":["begin","open","high","low","close"],"data":[["2026-03-01 10:00:00",10,12,null,11]]}}`)
	})

	rows, err := client.Chart(context.Background(), sampleHistory())
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 1 {
		t.Fatalf("rows = %+v", rows)
	}
	row := rows[0]
	if row.Open.String() != "10" || row.Close.String() != "11" || row.High == nil || row.High.String() != "12" || row.Low != nil {
		t.Fatalf("row = %+v", row)
	}
	if row.Begin.Format(timeLayout) != "2026-03-01 10:00:00" {
		t.Fatalf("begin = %s", row.Begin)
	}
}

func TestChartEmpty(t *testing.T) {
	t.Parallel()
	client := newClient(t, func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.WriteString(w, `{"candles":{"columns":["begin","open","high","low","close"],"data":[]}}`)
	})
	rows, err := client.Chart(context.Background(), sampleHistory())
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 0 {
		t.Fatalf("rows = %+v", rows)
	}
}
