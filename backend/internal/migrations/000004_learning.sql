-- material read progress
create table if not exists material_reads (
                                              user_id uuid not null,
                                              material_id uuid not null references materials(id) on delete cascade,
    group_id uuid not null references groups(id) on delete cascade,
    read_at timestamptz not null default now(),
    primary key (user_id, material_id)
    );
create index if not exists idx_material_reads_group on material_reads(group_id, user_id);

-- assignments
create table if not exists assignments (
                                           id uuid primary key,
                                           group_id uuid not null references groups(id) on delete cascade,
    title text not null,
    description text not null default '',
    due_at timestamptz null,
    created_by_user_id uuid not null,
    created_at timestamptz not null default now(),
    updated_at timestamptz not null default now()
    );
create index if not exists idx_assignments_group on assignments(group_id, created_at desc);

-- submissions (one per student per assignment for MVP)
create table if not exists submissions (
                                           id uuid primary key,
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

-- review (history allowed; latest review is what we show)
create table if not exists submission_reviews (
                                                  id uuid primary key,
                                                  submission_id uuid not null references submissions(id) on delete cascade,
    reviewer_user_id uuid not null,
    grade int null, -- MVP: optional
    comment text not null default '',
    created_at timestamptz not null default now()
    );
create index if not exists idx_reviews_submission on submission_reviews(submission_id, created_at desc);
