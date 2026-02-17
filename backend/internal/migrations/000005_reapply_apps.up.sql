-- 1) убрать старый constraint (он называется именно так у тебя в ошибке)
alter table enrollment_applications
drop constraint if exists enrollment_applications_user_id_group_id_key;

-- 2) вместо него частичный unique на активные заявки
create unique index if not exists ux_enrollment_apps_user_group_active
    on enrollment_applications (user_id, group_id)
    where status in ('submitted', 'in_review', 'approved');
