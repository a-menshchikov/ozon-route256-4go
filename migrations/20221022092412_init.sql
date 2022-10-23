-- +goose Up
-- +goose StatementBegin
create table users
(
  id serial primary key
);

create table tg_users
(
  id      int primary key,
  user_id int references users
);

create table expenses
(
  id            serial primary key,
  user_id       int references users,
  date          date       not null,
  amount        bigint     not null,
  currency_code varchar(3) not null,
  category      text       not null
);

create table currencies
(
  user_id int references users primary key,
  code    varchar(3) not null
);

create table limits
(
  user_id       int references users,
  category      text,
  total         bigint     not null,
  remains       bigint     not null,
  currency_code varchar(3) not null,

  primary key (user_id, category)
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
