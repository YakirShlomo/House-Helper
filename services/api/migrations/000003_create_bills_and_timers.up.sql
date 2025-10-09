-- Create bills table
CREATE TABLE IF NOT EXISTS bills (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    description TEXT,
    category VARCHAR(100) NOT NULL,
    amount DECIMAL(10,2) NOT NULL,
    currency VARCHAR(10) NOT NULL DEFAULT 'USD',
    due_date TIMESTAMP NOT NULL,
    is_recurring BOOLEAN DEFAULT FALSE,
    recurrence_rule TEXT,
    status VARCHAR(20) NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'paid', 'overdue', 'cancelled')),
    household_id UUID NOT NULL REFERENCES households(id) ON DELETE CASCADE,
    assigned_to UUID REFERENCES users(id),
    created_by UUID NOT NULL REFERENCES users(id),
    reminder_days INTEGER DEFAULT 3,
    auto_pay_enabled BOOLEAN DEFAULT FALSE,
    payment_method VARCHAR(100),
    vendor_info JSONB,
    attachment_urls TEXT[],
    paid_at TIMESTAMP,
    paid_by UUID REFERENCES users(id),
    paid_amount DECIMAL(10,2),
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP
);

-- Create bill payments table
CREATE TABLE IF NOT EXISTS bill_payments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    bill_id UUID NOT NULL REFERENCES bills(id) ON DELETE CASCADE,
    amount DECIMAL(10,2) NOT NULL,
    payment_method VARCHAR(100) NOT NULL,
    transaction_id VARCHAR(255),
    paid_by UUID NOT NULL REFERENCES users(id),
    paid_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    notes TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP
);

-- Create timers table
CREATE TABLE IF NOT EXISTS timers (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    description TEXT,
    category VARCHAR(100) NOT NULL,
    type VARCHAR(20) NOT NULL CHECK (type IN ('countdown', 'stopwatch', 'pomodoro')),
    duration INTEGER, -- in seconds
    household_id UUID NOT NULL REFERENCES households(id) ON DELETE CASCADE,
    created_by UUID NOT NULL REFERENCES users(id),
    status VARCHAR(20) NOT NULL DEFAULT 'created' CHECK (status IN ('created', 'running', 'paused', 'completed', 'stopped')),
    workflow_id VARCHAR(255), -- Temporal workflow ID
    settings JSONB,
    started_at TIMESTAMP,
    completed_at TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP
);

-- Create timer sessions table
CREATE TABLE IF NOT EXISTS timer_sessions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    timer_id UUID NOT NULL REFERENCES timers(id) ON DELETE CASCADE,
    started_at TIMESTAMP NOT NULL,
    ended_at TIMESTAMP,
    duration INTEGER, -- in seconds
    paused_duration INTEGER DEFAULT 0, -- in seconds
    notes TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP
);

-- Create indexes for bills
CREATE INDEX IF NOT EXISTS idx_bills_household_id ON bills(household_id);
CREATE INDEX IF NOT EXISTS idx_bills_assigned_to ON bills(assigned_to);
CREATE INDEX IF NOT EXISTS idx_bills_created_by ON bills(created_by);
CREATE INDEX IF NOT EXISTS idx_bills_paid_by ON bills(paid_by);
CREATE INDEX IF NOT EXISTS idx_bills_status ON bills(status);
CREATE INDEX IF NOT EXISTS idx_bills_category ON bills(category);
CREATE INDEX IF NOT EXISTS idx_bills_due_date ON bills(due_date);
CREATE INDEX IF NOT EXISTS idx_bills_is_recurring ON bills(is_recurring);
CREATE INDEX IF NOT EXISTS idx_bills_deleted_at ON bills(deleted_at);

-- Create indexes for bill payments
CREATE INDEX IF NOT EXISTS idx_bill_payments_bill_id ON bill_payments(bill_id);
CREATE INDEX IF NOT EXISTS idx_bill_payments_paid_by ON bill_payments(paid_by);
CREATE INDEX IF NOT EXISTS idx_bill_payments_paid_at ON bill_payments(paid_at);
CREATE INDEX IF NOT EXISTS idx_bill_payments_deleted_at ON bill_payments(deleted_at);

-- Create indexes for timers
CREATE INDEX IF NOT EXISTS idx_timers_household_id ON timers(household_id);
CREATE INDEX IF NOT EXISTS idx_timers_created_by ON timers(created_by);
CREATE INDEX IF NOT EXISTS idx_timers_status ON timers(status);
CREATE INDEX IF NOT EXISTS idx_timers_type ON timers(type);
CREATE INDEX IF NOT EXISTS idx_timers_category ON timers(category);
CREATE INDEX IF NOT EXISTS idx_timers_workflow_id ON timers(workflow_id);
CREATE INDEX IF NOT EXISTS idx_timers_deleted_at ON timers(deleted_at);

-- Create indexes for timer sessions
CREATE INDEX IF NOT EXISTS idx_timer_sessions_timer_id ON timer_sessions(timer_id);
CREATE INDEX IF NOT EXISTS idx_timer_sessions_started_at ON timer_sessions(started_at);
CREATE INDEX IF NOT EXISTS idx_timer_sessions_deleted_at ON timer_sessions(deleted_at);

-- Create triggers for updated_at
CREATE TRIGGER update_bills_updated_at BEFORE UPDATE ON bills FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_bill_payments_updated_at BEFORE UPDATE ON bill_payments FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_timers_updated_at BEFORE UPDATE ON timers FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_timer_sessions_updated_at BEFORE UPDATE ON timer_sessions FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();