-- +goose Up
-- +goose StatementBegin
create index if not exists idx_expenses_date_user on expenses (date, user_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
drop index idx_expenses_date_user;
-- +goose StatementEnd
