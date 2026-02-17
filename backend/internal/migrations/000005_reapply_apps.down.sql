drop index if exists ux_enrollment_apps_user_group_active;

alter table enrollment_applications
    add constraint enrollment_applications_user_id_group_id_key unique (user_id, group_id);
