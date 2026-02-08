create extension if not exists "uuid-ossp";

-- catalog
create table programs (
                          id uuid primary key,
                          title text not null,
                          description text not null default '',
                          created_at timestamptz not null default now()
);

create table cohorts (
                         id uuid primary key,
                         program_id uuid not null references programs(id) on delete cascade,
                         year int not null,
                         created_at timestamptz not null default now(),
                         unique(program_id, year)
);

create table groups (
                        id uuid primary key,
                        program_id uuid not null references programs(id) on delete cascade,
                        cohort_id uuid not null references cohorts(id) on delete cascade,
                        title text not null,
                        capacity int not null check (capacity >= 0),
                        created_at timestamptz not null default now()
);

-- applications
create table enrollment_applications (
                                         id uuid primary key,
                                         user_id uuid not null,
                                         group_id uuid not null references groups(id) on delete cascade,
                                         status text not null,
                                         comment text not null default '',
                                         created_at timestamptz not null default now(),
                                         updated_at timestamptz not null default now(),
                                         unique(user_id, group_id)
);

create index idx_enroll_apps_group on enrollment_applications(group_id);
create index idx_enroll_apps_user on enrollment_applications(user_id);
create index idx_enroll_apps_status on enrollment_applications(status);

-- audit of status changes
create table application_status_audit (
                                          id uuid primary key,
                                          application_id uuid not null references enrollment_applications(id) on delete cascade,
                                          actor_user_id uuid not null,
                                          actor_role text not null,
                                          from_status text not null,
                                          to_status text not null,
                                          reason text not null default '',
                                          created_at timestamptz not null default now()
);

create index idx_audit_app on application_status_audit(application_id);

-- enrollments
create table enrollments (
                             id uuid primary key,
                             user_id uuid not null,
                             group_id uuid not null references groups(id) on delete cascade,
                             created_at timestamptz not null default now(),
                             unique(user_id, group_id)
);

create index idx_enroll_group on enrollments(group_id);

-- outbox (Event -> Rule -> Action)
create table outbox_events (
                               id uuid primary key,
                               aggregate_type text not null,
                               aggregate_id uuid not null,
                               event_type text not null,
                               payload jsonb not null,
                               created_at timestamptz not null default now(),
                               published_at timestamptz null
);

create index idx_outbox_unpublished on outbox_events(published_at) where published_at is null;
