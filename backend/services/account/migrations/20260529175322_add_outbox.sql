-- +goose Up
create table if not exists outbox
(
    id bigint not null primary key ,
    aggregate_id bigint not null,
    aggregate_type varchar(50) not null,
    event_type varchar(50) not null,
    payload jsonb,
    created_at timestamp with time zone not null,
    published_at timestamp with time zone,
    trace_id varchar(255)
);

create index if not exists idx_published
    on outbox (published_at, created_at);

-- +goose Down
drop table if exists outbox cascade;
