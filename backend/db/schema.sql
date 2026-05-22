create schema if not exists account_service;

drop table if exists account_service.accounts CASCADE;

create table account_service.accounts
(
    id bigint not null
        constraint pk_accounts
            primary key,

    username varchar(50) not null
        constraint uq_accounts_username
            unique,

    password_hash varchar(255) not null,

    password_change_required boolean default false not null,

    display_name varchar(50) not null,

    role varchar(20) default 'member' not null
        constraint chk_accounts_role
            check (role in ('admin', 'moderator', 'member')),

    status varchar(20) default 'active' not null
        constraint chk_accounts_status
            check (status in ('active', 'suspended', 'archived')),

    created_at timestamp with time zone default current_timestamp not null,

    updated_at timestamp with time zone default current_timestamp not null,

    last_login_at timestamp with time zone,

    last_login_ip inet
);

create index idx_accounts_status on account_service.accounts (status);
create index idx_accounts_role on account_service.accounts (role);
