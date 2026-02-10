-- reverse of 000004_learning.up.sql

drop index if exists idx_reviews_submission;
drop table if exists submission_reviews;

drop index if exists idx_submissions_student;
drop index if exists idx_submissions_group;
drop table if exists submissions;

drop index if exists idx_assignments_group;
drop table if exists assignments;

drop index if exists idx_material_reads_group;
drop table if exists material_reads;
