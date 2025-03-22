create extension if not exists "uuid-ossp";
create extension if not exists "intarray";

create type API_STATUS as enum ('loading', 'online');

-- При запуске пода будут следующие значения
-- current_generation = null
-- next_generation = номер последней генерации данных
-- status = loading
-- last_activity = будет обновляться каждые 30 секунд если под живой

-- После запуска пода и загрузки последней генерации будут следующие значения
-- current_generation = номер последней генерации данных
-- next_generation = null
-- status = ready
-- last_activity = будет обновляться каждые 30 секунд если под живой

-- Если нужно будет загрузить новую генерацию в память, то нужно будет
-- выставить в базе данных следующие значения:
-- next_generation = номер новой генерации данных
-- Каждые 30 секунд поды будут обращаться к базе данных и сравнивать
-- номер генерации которая сейчас в памяти, и которая выставлена в поле
-- next_generation. Если next_generation более свежий, то под выставляет
-- status = loading и начинает загрузку свежей генерации в память.
-- После загрузки свежей генерации в память, под будет выставлять в
-- next_generation = null, а в поля current_generation новую загруженную версию и status = online

create table if not exists "public"."pods_states" (
    current_generation timestamptz default null,
    next_generation timestamptz default null,
    status API_STATUS not null,
    hostname text not null,
    last_activity timestamptz not null
);

create table if not exists "public"."seo_declarations" (
    generation timestamptz not null,
    url text not null,
    meta_title text default null,
    meta_description text default null,
    meta_robots text default null,
    meta_keywords text default null,
    faq jsonb not null default '{}'::jsonb,
    tags_cloud jsonb not null default '{}'::jsonb,
    created_at timestamptz default CURRENT_TIMESTAMP,
    updated_at timestamptz default CURRENT_TIMESTAMP,
    primary key ("generation", "url")
);
create index seo_declarations_generation_idx on "public"."seo_declarations" ("generation");
