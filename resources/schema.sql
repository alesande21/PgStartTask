drop schema if exists Script cascade;

create schema if not exists Script;

CREATE TYPE command_status AS ENUM ('aborted', 'crush', 'ended', 'in_progress', 'new');

create table Script.scripts (
    id serial primary key,
    body_script varchar not null,
    result_run_script varchar,
    status command_status
)