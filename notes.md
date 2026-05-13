# Stocks calculator project

Description

This project is designed to help me manage my stocks. Fractions of each company's stocks in relation to total portfolio cost change over some period, so next time I want to deposit money (or just rebalance portfolio) I need to know which stocks to buy or sell to preserve original ratios.
So app has functionality of adding groups of stocks that represents industrial fields. User is able to set ratios between these groups and ratios between individual stocks inside each group. Using whese ratios and current stock prices program will present information on what to buy/sell to rebalance portfolio.

server's options can be changed in config.yaml or in .env file
user can have multiple profiles, each profile can have multiple groups, groups can have multiple stocks.
rebalance function devides value by groups' weights and within each group devides equaly amoung all stocks

Stucture of the project

Project should be devided to 2 parts:
1. cli application that user will interact with
2. backend service that stores stocks info in db or gets it from the internet and deliveres it to users

This division makes it possible to create an account system, so I can access same database from multiple devices (and even have multiple accounts)


Files:
apitest  cmd  compose.yaml  config.yaml  Dockerfile  docs  go.mod  go.sum  internal  migrations  mocks  notes.md  pkg

./apitest:
commands  price.json  resonse.json  response.json  sber.json  spec.json  test2.json  test.json  trades.json  yandex.json

./cmd:
cli  server

./cmd/cli:
main.go

./cmd/server:
main.go

./docs:
docs.go  swagger.json  swagger.yaml

./internal:
cli  model  server

./internal/cli:
app  client  commands  config

./internal/cli/app:
connect.go  group.go  login.go  profile.go  rebalance.go  stock.go  terminal.go  validate_config.go

./internal/cli/client:
auth.go  client.go  group.go  helper.go  profile.go  rebalance.go  stock.go

./internal/cli/commands:
connect  group  login  profile  rebalance  root.go  stock

./internal/cli/commands/connect:
connect.go

./internal/cli/commands/group:
add_group.go  group.go  remove_group.go  rename_group.go  weight_group.go

./internal/cli/commands/login:
login.go

./internal/cli/commands/profile:
profile.go

./internal/cli/commands/rebalance:
rebalance.go

./internal/cli/commands/stock:
add_stock.go  remove_stock.go  stock.go

./internal/cli/config:
manage_config.go  user_info.go

./internal/model:
group.go  login.go  moex.go  profile.go  rebalance.go  stock.go

./internal/server:
app  config  config.go  handler  model  moex  repo  servererrors  server.go

./internal/server/app:
app.go  auth.go  group.go  profile.go  rebalance.go  rebalance_helper.go  rebalance_test.go  stock.go

./internal/server/config:

./internal/server/handler:
auth.go  group.go  handler.go  helper.go  middleware.go  profile.go  rebalance.go  stock.go  swagger_debug.go  swagger_release.go

./internal/server/model:

./internal/server/moex:
find.go  moex.go  price.go

./internal/server/repo:
auth.go  group.go  helper.go  profile.go  repo.go  stock.go  token.go

./internal/server/servererrors:
errors.go

./migrations:
000001_create_users.down.sql     000003_create_user_refresh_token.down.sql   000005_create_groups.down.sql  000007_create_transactions.down.sql
000001_create_users.up.sql       000003_create_user_refresh_token.up.sql     000005_create_groups.up.sql    000007_create_transactions.up.sql
000002_create_profiles.down.sql  000004_add_constraint_to_profiles.down.sql  000006_create_stocks.down.sql  notes
000002_create_profiles.up.sql    000004_add_constraint_to_profiles.up.sql    000006_create_stocks.up.sql

./mocks:
App.go  MoexApi.go  Repository.go  TokenManager.go

./pkg:
tokenmanager

./pkg/tokenmanager:
token_manager.go
