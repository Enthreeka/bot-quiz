set timezone = 'Europe/Moscow';

DO $$
    BEGIN
        IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'role') THEN
            CREATE TYPE role AS ENUM ('user', 'admin','superAdmin');
        END IF;
END $$;


DO $$
    BEGIN
        IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'chan_status') THEN
            CREATE TYPE chan_status AS ENUM ('kicked','administrator','left','member','unknown');
        END IF;
    END $$;


create table if not exists "user"
(
    id           bigint unique,
    tg_username  text                not null,
    created_at   timestamp           not null,
    channel_from varchar(150)        null,
    user_role    role default 'user' not null,
    primary key (id)
);

create table if not exists channel(
                                      id int generated always as identity,
                                      tg_id bigint unique not null,
                                      channel_name varchar(150) null,
                                      channel_url varchar(150) null,
                                      channel_status chan_status not null,
                                      primary key (id)
);

create table if not exists questions(
                                        id int generated always as identity,
                                        created_by_user bigint,
                                        created_at timestamp with time zone default now(),
                                        question_name text,
                                        file_id varchar(100),
                                        deadline timestamp,
                                        is_send boolean default false,
                                        primary key (id)
);


create table if not exists answers(
    id int generated always as identity,
    answer varchar(100),
    cost_of_response int,
    question_id int,
    foreign key (question_id)
        references questions (id) on delete cascade,
    primary key (id)
);

create table if not exists user_results(
                                           id int generated always as identity,
                                           user_id bigint,
                                           points int,
                                           questions_id int,
                                           primary key (id),
                                           foreign key (user_id)
                                               references "user" (id) on delete cascade,
                                           foreign key (questions_id)
                                               references questions (id)
);

create table if not exists is_user_answer(
    user_id bigint,
    is_answer boolean default false,
    question_id int not null,
    foreign key (user_id)
        references "user" (id) on delete cascade,
    foreign key (question_id)
        references questions (id) on delete cascade
);



-- new version
alter table questions add column channel_tg_id bigint ;

ALTER TABLE questions
    ADD CONSTRAINT fk_channel
        FOREIGN KEY (channel_tg_id) REFERENCES channel(tg_id) on delete cascade;


alter table user_results add column questions_id int;

ALTER TABLE user_results
    ADD CONSTRAINT fk_questions
        FOREIGN KEY (questions_id) REFERENCES questions(id) on delete cascade;


ALTER TABLE user_results
    rename column total_points to points;

SELECT constraint_name
FROM information_schema.table_constraints
WHERE table_name = 'user_results' AND constraint_type = 'UNIQUE';

ALTER TABLE user_results
        DROP CONSTRAINT user_results_user_id_key;




