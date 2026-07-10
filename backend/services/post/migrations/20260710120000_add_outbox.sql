-- +goose Up
create table if not exists outbox
(
    id             bigint                                                   not null
        constraint pk_outbox
            primary key,
    aggregate_id   bigint                                                   not null,
    aggregate_type text                                                     not null,
    topic          text                                                     not null,
    event_type     text                                                     not null,
    payload        jsonb,
    created_at     timestamp with time zone default CURRENT_TIMESTAMP        not null,
    published_at   timestamp with time zone,
    trace_id       text
);

create index if not exists idx_outbox_unpublished
    on outbox (created_at)
    where published_at is null;

-- +goose Down
drop table if exists outbox cascade;
