-- program visibility
alter table programs add column status text not null default 'draft';
create index if not exists idx_programs_status on programs(status);

-- group flags
alter table groups add column is_open boolean not null default true;
alter table groups add column requires_interview boolean not null default true;

-- teacher assignment (teacher is an assignment, not global role)
create table if not exists group_teachers (
                                              group_id uuid not null references groups(id) on delete cascade,
    teacher_user_id uuid not null,
    created_at timestamptz not null default now(),
    primary key (group_id, teacher_user_id)
    );
create index if not exists idx_group_teachers_teacher on group_teachers(teacher_user_id);

-- interviews (1 interview per application for MVP)
create table if not exists interviews (
                                          id uuid primary key,
                                          application_id uuid not null references enrollment_applications(id) on delete cascade,
    group_id uuid not null references groups(id) on delete cascade,
    candidate_user_id uuid not null,
    interviewer_user_id uuid not null,
    interviewer_role text not null, -- 'teacher'|'moderator'
    result text not null,           -- pending|recommended|not_recommended|needs_more
    comment text not null default '',
    created_at timestamptz not null default now(),
    updated_at timestamptz not null default now(),
    unique(application_id)
    );
create index if not exists idx_interviews_group on interviews(group_id);
