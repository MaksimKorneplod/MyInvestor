package quotes

import "github.com/MaksimKorneplod/MyInvestor/internal/repository/moex"

// Instrument и Quote — значения репозитория MOEX, которые сервис отдаёт наружу.
// HTTP-слой зависит от этих имён и не импортирует repository.
type (
	Instrument = moex.Instrument
	Quote      = moex.Quote
)
