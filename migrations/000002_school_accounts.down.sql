BEGIN;
DROP TABLE favorite_stocks;
DROP TRIGGER student_create_info ON student;
DROP FUNCTION create_student_info();
DROP TABLE student_info;
DROP TABLE student;
DROP TABLE school;
COMMIT;
