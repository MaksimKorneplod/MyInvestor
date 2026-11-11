BEGIN;

CREATE TABLE school (
    school_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    login VARCHAR(50) NOT NULL CHECK (length(btrim(login)) > 0),
    email VARCHAR(254) NOT NULL CHECK (length(btrim(email)) > 0),
    password_hash VARCHAR(255) NOT NULL CHECK (length(password_hash) > 0)
);
CREATE UNIQUE INDEX school_email_unique ON school (lower(email));
CREATE UNIQUE INDEX school_login_unique ON school (lower(login));

CREATE TABLE student (
    student_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    school_id UUID NOT NULL REFERENCES school(school_id),
    login VARCHAR(50) NOT NULL CHECK (length(btrim(login)) > 0),
    email VARCHAR(254) NOT NULL CHECK (length(btrim(email)) > 0),
    password_hash VARCHAR(255) NOT NULL CHECK (length(password_hash) > 0),
    entry_date DATE NOT NULL DEFAULT CURRENT_DATE,
    expiration_date DATE NOT NULL,
    CHECK (expiration_date >= entry_date)
);
CREATE INDEX student_school_id_idx ON student(school_id);
CREATE UNIQUE INDEX student_email_unique ON student(lower(email));
CREATE UNIQUE INDEX student_login_unique ON student(lower(login));

-- Shared primary key enforces at most one profile per student.
-- The trigger below creates the mandatory profile in the same transaction.
CREATE TABLE student_info (
    student_id UUID PRIMARY KEY REFERENCES student(student_id) ON DELETE CASCADE,
    balance BIGINT NOT NULL DEFAULT 100000 CHECK (balance >= 0),
    streak_count INTEGER NOT NULL DEFAULT 0 CHECK (streak_count >= 0),
    streak_last_active_date TIMESTAMPTZ
);
COMMENT ON COLUMN student_info.balance IS 'Minor currency units: 100000 = 1000.00';

CREATE FUNCTION create_student_info() RETURNS trigger LANGUAGE plpgsql AS $$
BEGIN
    INSERT INTO student_info(student_id) VALUES (NEW.student_id);
    RETURN NEW;
END;
$$;
CREATE TRIGGER student_create_info AFTER INSERT ON student
    FOR EACH ROW EXECUTE FUNCTION create_student_info();

CREATE TABLE favorite_stocks (
    student_id UUID NOT NULL REFERENCES student(student_id) ON DELETE CASCADE,
    stock_symbol VARCHAR(20) NOT NULL CHECK (length(btrim(stock_symbol)) > 0),
    PRIMARY KEY (student_id, stock_symbol)
);

COMMIT;
