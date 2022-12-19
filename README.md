# simplebank

## 7. Fix Deadlock

### Cause

`account id` foreign key constraint in `transfers` table

### Solution

use `select for no key update`

## 8. How To Avoid Deadlock in DB Transaction?

- Queries orders matters

## To Do

- Generate open api 3 doc using grpc plugin
- Need to search grpc plugin support open api 3 doc
- add `make help` command
