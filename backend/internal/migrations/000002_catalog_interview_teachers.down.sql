drop table if exists interviews;
drop table if exists group_teachers;

alter table groups drop column if exists requires_interview;
alter table groups drop column if exists is_open;

alter table programs drop column if exists status;
