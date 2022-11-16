-- +goose Up
-- +goose StatementBegin
create table users
(
  id serial primary key
);

create table tg_users
(
  id      int,
  user_id int not null,

  primary key (id),
  foreign key (user_id) references users
    on delete cascade
);

create table expenses
(
  id            serial,
  user_id       int        not null,
  date          date       not null,
  amount        bigint     not null,
  currency_code varchar(3) not null,
  category      text       not null,

  primary key (id),
  foreign key (user_id) references users
    on delete cascade
);

create table currencies
(
  user_id int,
  code    varchar(3) not null,

  primary key (user_id),
  foreign key (user_id) references users
    on delete cascade
);

create table limits
(
  user_id       int,
  category      text,
  total         bigint     not null,
  remains       bigint     not null,
  currency_code varchar(3) not null,

  primary key (user_id, category),
  foreign key (user_id) references users
    on delete cascade
);

create table rates
(
  code varchar(3),
  date date,
  rate bigint not null,

  primary key (code, date)
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
drop table rates;
drop table limits;
drop table currencies;
drop table expenses;
drop table tg_users;
drop table users;
-- +goose StatementEnd
