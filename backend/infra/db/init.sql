-- Database users
--   migration tool user (goose)
create user goose with password 'goose_password';

--   account service runtime user
create user account_service_svc with password 'account_service_svc_password';

--   post service runtime user
create user post_service_svc with password 'post_service_svc_password';

--   openfga authorization service user
create user openfga with password 'openfga_dev_password';


-- Database connect privileges
grant connect on database community to goose;
grant connect on database community to account_service_svc;
grant connect on database community to post_service_svc;
grant connect on database community to openfga;


-- Application schemas
create schema if not exists account_service;
create schema if not exists post_service;
create schema if not exists openfga;


-- Schema-level usage & creation privileges
--   account_service schema:
--     goose: full management rights (can create objects inside the schema)
--     account_service_svc: runtime access only (usage)
grant usage, create on schema account_service to goose;
grant usage on schema account_service to account_service_svc;

grant usage, create on schema post_service to goose;
grant usage on schema post_service to post_service_svc;

grant usage, create on schema openfga to openfga;


-- Existing table & sequence access
--   give account_service_svc full DML on all current account_service tables
grant select, insert, update, delete on all tables in schema account_service to account_service_svc;
grant select, insert, update, delete on all tables in schema post_service to post_service_svc;
--   and read access to current sequences
grant usage, select on all sequences in schema account_service to account_service_svc;
grant usage, select on all sequences in schema post_service to post_service_svc;


-- Default privileges for future objects

--   account_service – objects created by community (goose runs as community by default)
alter default privileges for role community in schema account_service
    grant select, insert, update, delete on tables to account_service_svc;
alter default privileges for role community in schema account_service
    grant all on tables to goose;
alter default privileges for role community in schema account_service
    grant usage on sequences to account_service_svc;
alter default privileges for role community in schema account_service
    grant all on sequences to goose;
alter default privileges for role community in schema account_service
    grant execute on functions to goose;

--   account_service – objects created directly by goose must also be accessible by the runtime user
alter default privileges for role goose in schema account_service
    grant select, insert, update, delete on tables to account_service_svc;
alter default privileges for role goose in schema account_service
    grant usage, select on sequences to account_service_svc;

--   post_service - objects created by community (goose runs as community by default)
alter default privileges for role community in schema post_service
    grant select, insert, update, delete on tables to post_service_svc;
alter default privileges for role community in schema post_service
    grant all on tables to goose;
alter default privileges for role community in schema post_service
    grant usage on sequences to post_service_svc;
alter default privileges for role community in schema post_service
    grant all on sequences to goose;
alter default privileges for role community in schema post_service
    grant execute on functions to goose;

--   post_service – objects created directly by goose must also be accessible by the runtime user
alter default privileges for role goose in schema post_service
    grant select, insert, update, delete on tables to post_service_svc;
alter default privileges for role goose in schema post_service
    grant usage, select on sequences to post_service_svc;

--   openfga – openfga user owns everything in its schema
alter default privileges for role community in schema openfga
    grant all on tables to openfga;
alter default privileges for role community in schema openfga
    grant all on sequences to openfga;