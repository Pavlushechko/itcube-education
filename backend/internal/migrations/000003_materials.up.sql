create table if not exists materials (
                                         id uuid primary key,
                                         group_id uuid not null references groups(id) on delete cascade,
    type text not null, -- file|link|text|video
    title text not null,
    content text not null default '', -- url/text/etc
    created_by_user_id uuid not null,
    created_at timestamptz not null default now()
    );

create index if not exists idx_materials_group on materials(group_id);
