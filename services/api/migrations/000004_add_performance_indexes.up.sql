-- Add performance indexes for common queries

-- User authentication and lookup indexes
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_users_email_verified ON users(email, is_verified) WHERE deleted_at IS NULL;
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_users_phone_verified ON users(phone, is_verified) WHERE deleted_at IS NULL;

-- Household member queries
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_household_members_active ON household_members(household_id, user_id) WHERE left_at IS NULL;
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_household_members_user_active ON household_members(user_id, household_id, role) WHERE left_at IS NULL;

-- Task filtering and sorting
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_tasks_household_status_priority ON tasks(household_id, status, priority) WHERE deleted_at IS NULL;
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_tasks_assigned_status_due ON tasks(assigned_to, status, due_date) WHERE deleted_at IS NULL;
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_tasks_overdue ON tasks(household_id, due_date) WHERE status = 'pending' AND due_date < CURRENT_TIMESTAMP AND deleted_at IS NULL;

-- Shopping list collaboration
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_shopping_items_list_purchased ON shopping_items(list_id, is_purchased) WHERE deleted_at IS NULL;
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_shopping_items_category_purchased ON shopping_items(list_id, category, is_purchased) WHERE deleted_at IS NULL;

-- Bill tracking and reminders
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_bills_household_status_due ON bills(household_id, status, due_date) WHERE deleted_at IS NULL;
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_bills_upcoming ON bills(household_id, due_date) WHERE status = 'pending' AND due_date BETWEEN CURRENT_TIMESTAMP AND CURRENT_TIMESTAMP + INTERVAL '30 days' AND deleted_at IS NULL;
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_bills_overdue ON bills(household_id, due_date) WHERE status = 'pending' AND due_date < CURRENT_TIMESTAMP AND deleted_at IS NULL;

-- Timer status tracking
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_timers_household_status ON timers(household_id, status) WHERE deleted_at IS NULL;
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_timers_active ON timers(household_id, status, created_by) WHERE status IN ('running', 'paused') AND deleted_at IS NULL;

-- Invitation management
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_household_invitations_pending ON household_invitations(household_id, status, expires_at) WHERE status = 'pending' AND deleted_at IS NULL;

-- Product search optimization
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_products_name_trgm ON products USING gin(name gin_trgm_ops);
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_products_brand_category ON products(brand, category);

-- Composite indexes for common filter combinations
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_tasks_category_priority_status ON tasks(household_id, category, priority, status) WHERE deleted_at IS NULL;
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_bills_category_status_due ON bills(household_id, category, status, due_date) WHERE deleted_at IS NULL;