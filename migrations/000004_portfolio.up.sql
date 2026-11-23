BEGIN;

-- References the same stable MOEX instrument ID that the catalog returns.
CREATE TABLE trading_instrument (
    isin VARCHAR(12) PRIMARY KEY CHECK (isin ~ '^[A-Z]{2}[A-Z0-9]{9}[0-9]$')
);

CREATE TABLE portfolio_position (
    student_id UUID NOT NULL REFERENCES student(student_id) ON DELETE CASCADE,
    instrument_isin VARCHAR(12) NOT NULL REFERENCES trading_instrument(isin),
    quantity BIGINT NOT NULL CHECK (quantity > 0),
    average_price NUMERIC(24, 8) NOT NULL CHECK (average_price > 0 AND average_price <> 'NaN'),
    PRIMARY KEY (student_id, instrument_isin)
);

-- Closing a position deletes its row. Trade history references the instrument
-- and student directly, so it remains even when there is no open position.
CREATE TABLE trade_operation (
    trade_id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    student_id UUID NOT NULL REFERENCES student(student_id) ON DELETE CASCADE,
    instrument_isin VARCHAR(12) NOT NULL REFERENCES trading_instrument(isin),
    side TEXT NOT NULL CHECK (side IN ('buy', 'sell')),
    quantity BIGINT NOT NULL CHECK (quantity > 0),
    price NUMERIC(24, 8) NOT NULL CHECK (price > 0 AND price <> 'NaN'),
    executed_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX trade_operation_student_history_idx
    ON trade_operation (student_id, executed_at DESC, trade_id DESC);

COMMENT ON COLUMN portfolio_position.average_price IS 'Average purchase price per share in rubles';
COMMENT ON COLUMN trade_operation.price IS 'Executed price per share in rubles';

COMMIT;
