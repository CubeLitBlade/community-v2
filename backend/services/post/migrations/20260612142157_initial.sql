-- +goose Up
create table if not exists posts
(
    id         bigint                                                          not null
        constraint pk_posts
            primary key,
    author_id  bigint                                                          not null,
    title      varchar(255),
    content    text                                                            not null,
    status     varchar(20)              default 'published'::character varying not null
        constraint chk_post_status
            check ((status)::text = ANY
                   ((ARRAY ['published'::character varying, 'archived'::character varying])::text[])),
    created_at timestamp with time zone default CURRENT_TIMESTAMP              not null,
    updated_at timestamp with time zone default CURRENT_TIMESTAMP              not null
);

create index if not exists idx_author_id
    on posts (author_id);

create index if not exists idx_posts_status
    on posts (status);

create table if not exists account_profiles
(
    id           bigint      not null
        constraint pk_account_profiles
            primary key,
    username     varchar(50) not null
        constraint uq_account_profiles_username
            unique,
    display_name varchar(50) not null
);

-- +goose Down
drop table if exists posts cascade;
drop table if exists account_profiles cascade;
