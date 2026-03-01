alter table groups add column if not exists teacher_id uuid;

-- если ранее использовалась group_teachers:
update groups g
set teacher_id = gt.teacher_user_id
from group_teachers gt
where gt.group_id = g.id
  and g.teacher_id is null;

create index if not exists idx_groups_teacher_id on groups(teacher_id);