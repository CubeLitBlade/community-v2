-- +goose Up
create table if not exists accounts
(
    id                       bigint                                                       not null
        constraint pk_accounts
            primary key,
    username                 varchar(50)                                                  not null
        constraint uq_accounts_username
            unique,
    password_hash            varchar(255)                                                 not null,
    password_change_required boolean                  default false                       not null,
    display_name             varchar(50)                                                  not null,
    role                     varchar(20)              default 'member'::character varying not null
        constraint chk_accounts_role
            check ((role)::text = ANY
                   ((ARRAY ['admin'::character varying, 'moderator'::character varying, 'member'::character varying])::text[])),
    status                   varchar(20)              default 'active'::character varying not null
        constraint chk_accounts_status
            check ((status)::text = ANY
                   ((ARRAY ['active'::character varying, 'suspended'::character varying, 'archived'::character varying])::text[])),
    created_at               timestamp with time zone default CURRENT_TIMESTAMP           not null,
    updated_at               timestamp with time zone default CURRENT_TIMESTAMP           not null,
    last_login_at            timestamp with time zone,
    last_login_ip            inet
);

create index if not exists idx_accounts_status
    on accounts (status);

create index if not exists idx_accounts_role
    on accounts (role);

-- +goose Down
drop table if exists accounts cascade;
