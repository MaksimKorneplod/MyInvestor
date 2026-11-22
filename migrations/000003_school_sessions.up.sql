BEGIN;

CREATE TABLE school_sessions (
    session_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    school_id UUID NOT NULL REFERENCES school(school_id) ON DELETE CASCADE,
    token_hash CHAR(64) NOT NULL UNIQUE,
    expires_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX school_sessions_school_id_idx ON school_sessions (school_id);
CREATE INDEX school_sessions_expires_at_idx ON school_sessions (expires_at);

COMMIT;
