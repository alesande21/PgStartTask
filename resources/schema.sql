drop schema if exists Script cascade;

create schema if not exists Script;

create table Script.scripts (
    id serial primary key,
    body_script varchar not null,
    result_run_script varchar,
    status int
)