CREATE TABLE IF NOT EXISTS public.purchases
(
    id serial primary key,
    purchaser varchar(255) not null,
    item varchar(255) not null,
    processed boolean not null default false,
    deleted boolean not null default false
);