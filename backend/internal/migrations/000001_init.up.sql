-- 000001_init.up.sql
-- Dev-friendly single init migration for education module

create extension if not exists pgcrypto;

-- =========================
-- Catalog
-- =========================
create table if not exists programs (
                                        id uuid primary key default gen_random_uuid(),
    title text not null,
    description text not null default '',
    status text not null default 'draft', -- draft|published|archived
    created_at timestamptz not null default now(),
    updated_at timestamptz not null default now()
    );
create index if not exists idx_programs_status on programs(status);

create table if not exists cohorts (
                                       id uuid primary key default gen_random_uuid(),
    program_id uuid not null references programs(id) on delete cascade,
    year int not null,
    created_at timestamptz not null default now(),
    unique(program_id, year)
    );

create table if not exists groups (
                                      id uuid primary key default gen_random_uuid(),
    program_id uuid not null references programs(id) on delete cascade,
    cohort_id uuid not null references cohorts(id) on delete cascade,
    title text not null,
    capacity int not null check (capacity >= 0),
    is_open boolean not null default true,
    requires_interview boolean not null default false,
    created_at timestamptz not null default now(),
    updated_at timestamptz not null default now()
    );

create index if not exists idx_groups_program_id on groups(program_id);
create index if not exists idx_groups_cohort_id on groups(cohort_id);
create index if not exists idx_groups_is_open on groups(is_open);

-- Teacher assignment (assignment, not global role)
-- NOTE: according to your latest requirement it's 1:1 group->teacher_id
-- BUT your current code uses group_teachers. Keeping it for now.
create table if not exists group_teachers (
                                              group_id uuid not null references groups(id) on delete cascade,
    teacher_user_id uuid not null,
    created_at timestamptz not null default now(),
    primary key (group_id, teacher_user_id)
    );
create index if not exists idx_group_teachers_teacher on group_teachers(teacher_user_id);

-- =========================
-- Applications + audit
-- =========================
create table if not exists enrollment_applications (
                                                       id uuid primary key default gen_random_uuid(),
    user_id uuid not null,
    group_id uuid not null references groups(id) on delete cascade,

    -- pending|interview|approved|rejected|enrolled
    status text not null default 'pending',

    comment text not null default '',
    created_at timestamptz not null default now(),
    updated_at timestamptz not null default now(),

    unique(user_id, group_id)
    );

create index if not exists idx_enroll_apps_group on enrollment_applications(group_id);
create index if not exists idx_enroll_apps_user on enrollment_applications(user_id);
create index if not exists idx_enroll_apps_status on enrollment_applications(status);

-- audit of status changes (who/when/comment)
create table if not exists application_status_audit (
                                                        id uuid primary key default gen_random_uuid(),
    application_id uuid not null references enrollment_applications(id) on delete cascade,
    actor_user_id uuid not null,
    actor_role text not null default 'user',
    from_status text not null,
    to_status text not null,
    reason text not null default '',
    created_at timestamptz not null default now()
    );
create index if not exists idx_audit_app on application_status_audit(application_id, created_at desc);

-- =========================
-- Enrollment (access to /learn/*)
-- =========================
create table if not exists enrollments (
                                           id uuid primary key default gen_random_uuid(),
    user_id uuid not null,
    group_id uuid not null references groups(id) on delete cascade,
    created_at timestamptz not null default now(),
    unique(user_id, group_id)
    );
create index if not exists idx_enroll_group on enrollments(group_id);
create index if not exists idx_enroll_user on enrollments(user_id);

-- =========================
-- Interviews (MVP: 1 per application)
-- =========================
create table if not exists interviews (
                                          id uuid primary key default gen_random_uuid(),
    application_id uuid not null references enrollment_applications(id) on delete cascade,
    group_id uuid not null references groups(id) on delete cascade,
    candidate_user_id uuid not null,
    interviewer_user_id uuid not null,
    interviewer_role text not null default 'teacher', -- teacher|moderator
    result text not null default 'pending',           -- pending|recommended|not_recommended|needs_more
    comment text not null default '',
    created_at timestamptz not null default now(),
    updated_at timestamptz not null default now(),
    unique(application_id)
    );
create index if not exists idx_interviews_group on interviews(group_id);

-- =========================
-- Outbox (Event -> Rule -> Action)
-- =========================
create table if not exists outbox_events (
                                             id uuid primary key default gen_random_uuid(),
    aggregate_type text not null,
    aggregate_id uuid not null,
    event_type text not null,
    payload jsonb not null,
    created_at timestamptz not null default now(),
    published_at timestamptz null
    );
create index if not exists idx_outbox_unpublished
    on outbox_events(published_at)
    where published_at is null;

-- =========================
-- Files (MinIO metadata)
-- =========================
create table if not exists files (
                                     id uuid primary key default gen_random_uuid(),
    bucket text not null default 'education',
    object_key text not null,
    original_name text not null,
    mime_type text not null default 'application/octet-stream',
    size_bytes bigint not null default 0,
    checksum_sha256 text,
    created_by_user_id uuid not null,
    created_at timestamptz not null default now(),
    unique(object_key)
    );
create index if not exists idx_files_created_by on files(created_by_user_id, created_at desc);

-- =========================
-- Materials + progress + attachments
-- =========================
create table if not exists materials (
                                         id uuid primary key default gen_random_uuid(),
    group_id uuid not null references groups(id) on delete cascade,
    type text not null, -- file|link|text|video
    title text not null,
    content text not null default '',
    external_url text null,
    created_by_user_id uuid not null,
    created_at timestamptz not null default now()
    );
create index if not exists idx_materials_group on materials(group_id, created_at desc);

-- material -> files (many-to-many)
create table if not exists material_files (
                                              material_id uuid not null references materials(id) on delete cascade,
    file_id uuid not null references files(id) on delete restrict,
    created_at timestamptz not null default now(),
    primary key (material_id, file_id)
    );
create index if not exists idx_material_files_material_id on material_files(material_id, created_at asc);
create index if not exists idx_material_files_file_id on material_files(file_id);

-- material read progress
create table if not exists material_reads (
                                              user_id uuid not null,
                                              material_id uuid not null references materials(id) on delete cascade,
    group_id uuid not null references groups(id) on delete cascade,
    read_at timestamptz not null default now(),
    primary key (user_id, material_id)
    );
create index if not exists idx_material_reads_group on material_reads(group_id, user_id);

-- =========================
-- Assignments + submissions + reviews
-- =========================
create table if not exists assignments (
                                           id uuid primary key default gen_random_uuid(),
    group_id uuid not null references groups(id) on delete cascade,
    title text not null,
    description text not null default '',
    due_at timestamptz null,
    is_key boolean not null default false,
    created_by_user_id uuid not null,
    created_at timestamptz not null default now(),
    updated_at timestamptz not null default now()
    );
create index if not exists idx_assignments_group on assignments(group_id, created_at desc);

create table if not exists submissions (
                                           id uuid primary key default gen_random_uuid(),
    assignment_id uuid not null references assignments(id) on delete cascade,
    group_id uuid not null references groups(id) on delete cascade,
    student_user_id uuid not null,
    content_type text not null,
    content text not null,
    status text not null default 'submitted', -- submitted|reviewed
    created_at timestamptz not null default now(),
    updated_at timestamptz not null default now(),
    unique (assignment_id, student_user_id)
    );
create index if not exists idx_submissions_group on submissions(group_id, status, created_at desc);
create index if not exists idx_submissions_student on submissions(student_user_id, created_at desc);

create table if not exists submission_reviews (
                                                  id uuid primary key default gen_random_uuid(),
    submission_id uuid not null references submissions(id) on delete cascade,
    reviewer_user_id uuid not null,
    grade int null,
    comment text not null default '',
    created_at timestamptz not null default now()
    );
create index if not exists idx_reviews_submission on submission_reviews(submission_id, created_at desc);