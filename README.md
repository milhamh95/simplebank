# simplebank

## 7. Fix Deadlock

### Cause

`account id` foreign key constraint in `transfers` table

### Solution

use `select for no key update`
