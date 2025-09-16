-- Drop tables in reverse order due to foreign key constraints
DROP TABLE IF EXISTS user_balances;
DROP TABLE IF EXISTS idempotency_keys;
DROP TABLE IF EXISTS settlements;
DROP TABLE IF EXISTS expense_splits;
DROP TABLE IF EXISTS expenses;
DROP TABLE IF EXISTS group_members;
DROP TABLE IF EXISTS `groups`;
DROP TABLE IF EXISTS users;
