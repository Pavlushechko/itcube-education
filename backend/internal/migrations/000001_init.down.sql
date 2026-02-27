-- 000001_init.down.sql
drop schema public cascade;
create schema public;
create extension if not exists pgcrypto;