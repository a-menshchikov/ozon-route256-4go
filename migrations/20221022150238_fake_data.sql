-- +goose Up
-- +goose StatementBegin
do
$$
    declare
        user_1 int;
        user_2 int;
    begin
        insert into users values (default) returning id into user_1;
        insert into users values (default) returning id into user_2;

        insert into tg_users (id, user_id)
        values (101, user_1),
               (102, user_2);

        insert into currencies (user_id, code)
        select unnest(array [user_2]),
               unnest(array ['USD']);

        insert into limits (user_id, category, total, remains, currency_code)
        select user_1, unnest(array ['', 'taxi']), unnest(array [25000000, 1000000]), unnest(array [20000000, 700000]), unnest(array ['RUB', 'USD']);

        insert into rates (code, date, rate)
        select 'USD', generate_series('2022-10-19'::timestamp, '2022-10-21'::timestamp, '1 day'), unnest(array [550000, 525000, 500000])
        union
        select 'EUR', generate_series('2022-10-19'::timestamp, '2022-10-21'::timestamp, '1 day'), unnest(array [650000, 625000, 600000]);

        insert into expenses (user_id, date, amount, currency_code, category)
        values (user_1, '2022-10-19', 100000, 'USD', 'taxi'),
               (user_1, '2022-10-20', 2500000, 'RUB', 'lunch'),
               (user_1, '2022-10-21', 10000000, 'RUB', 'taxi'),
               (user_1, '2022-10-21', 250000, 'USD', 'medicine'),
               (user_2, '2022-10-19', 1100000, 'RUB', 'coffee'),
               (user_2, '2022-10-20', 1575000, 'RUB', 'coffee'),
               (user_2, '2022-10-21', 30000, 'USD', 'coffee');
    end
$$ language plpgsql;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
truncate users cascade;
truncate rates;
-- +goose StatementEnd
