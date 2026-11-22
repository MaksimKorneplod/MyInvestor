package moex

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"
)

func TestListed(t *testing.T) {
	t.Parallel()
	client := newClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/iss/engines/stock/markets/shares/boards/TQBR/securities/SBER.json" {
			t.Errorf("path = %s", r.URL.Path)
		}
		if !strings.Contains(r.URL.Query().Get("securities.columns"), "ISIN") {
			t.Errorf("columns = %s", r.URL.Query().Get("securities.columns"))
		}
		if r.URL.Query().Get("start") != "0" {
			_, _ = io.WriteString(w, `{"securities":{"columns":["SECID","SHORTNAME","ISIN"],"data":[]}}`)
			return
		}
		_, _ = io.WriteString(w, `{"securities":{"columns":["SECID","SHORTNAME","ISIN"],"data":[["SBER","Сбербанк","RU0009029540"]]}}`)
	})
	found, err := client.Listed(context.Background(), "SBER")
	if err != nil {
		t.Fatal(err)
	}
	if !found {
		t.Fatal("SBER was not listed")
	}
}

func TestListedMissing(t *testing.T) {
	t.Parallel()
	client := newClient(t, func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.WriteString(w, `{"securities":{"columns":["SECID","SHORTNAME","ISIN"],"data":[]}}`)
	})
	found, err := client.Listed(context.Background(), "NOPE")
	if err != nil {
		t.Fatal(err)
	}
	if found {
		t.Fatal("NOPE is listed")
	}
}
