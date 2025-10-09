-- Drop triggers
DROP TRIGGER IF EXISTS update_timer_sessions_updated_at ON timer_sessions;
DROP TRIGGER IF EXISTS update_timers_updated_at ON timers;
DROP TRIGGER IF EXISTS update_bill_payments_updated_at ON bill_payments;
DROP TRIGGER IF EXISTS update_bills_updated_at ON bills;

-- Drop indexes
DROP INDEX IF EXISTS idx_timer_sessions_deleted_at;
DROP INDEX IF EXISTS idx_timer_sessions_started_at;
DROP INDEX IF EXISTS idx_timer_sessions_timer_id;
DROP INDEX IF EXISTS idx_timers_deleted_at;
DROP INDEX IF EXISTS idx_timers_workflow_id;
DROP INDEX IF EXISTS idx_timers_category;
DROP INDEX IF EXISTS idx_timers_type;
DROP INDEX IF EXISTS idx_timers_status;
DROP INDEX IF EXISTS idx_timers_created_by;
DROP INDEX IF EXISTS idx_timers_household_id;
DROP INDEX IF EXISTS idx_bill_payments_deleted_at;
DROP INDEX IF EXISTS idx_bill_payments_paid_at;
DROP INDEX IF EXISTS idx_bill_payments_paid_by;
DROP INDEX IF EXISTS idx_bill_payments_bill_id;
DROP INDEX IF EXISTS idx_bills_deleted_at;
DROP INDEX IF EXISTS idx_bills_is_recurring;
DROP INDEX IF EXISTS idx_bills_due_date;
DROP INDEX IF EXISTS idx_bills_category;
DROP INDEX IF EXISTS idx_bills_status;
DROP INDEX IF EXISTS idx_bills_paid_by;
DROP INDEX IF EXISTS idx_bills_created_by;
DROP INDEX IF EXISTS idx_bills_assigned_to;
DROP INDEX IF EXISTS idx_bills_household_id;

-- Drop tables in reverse order
DROP TABLE IF EXISTS timer_sessions;
DROP TABLE IF EXISTS timers;
DROP TABLE IF EXISTS bill_payments;
DROP TABLE IF EXISTS bills;