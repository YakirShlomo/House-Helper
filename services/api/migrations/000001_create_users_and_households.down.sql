-- Drop triggers
DROP TRIGGER IF EXISTS update_household_invitations_updated_at ON household_invitations;
DROP TRIGGER IF EXISTS update_household_members_updated_at ON household_members;
DROP TRIGGER IF EXISTS update_households_updated_at ON households;
DROP TRIGGER IF EXISTS update_user_profiles_updated_at ON user_profiles;
DROP TRIGGER IF EXISTS update_users_updated_at ON users;

-- Drop function
DROP FUNCTION IF EXISTS update_updated_at_column();

-- Drop indexes
DROP INDEX IF EXISTS idx_household_invitations_expires_at;
DROP INDEX IF EXISTS idx_household_invitations_status;
DROP INDEX IF EXISTS idx_household_invitations_email;
DROP INDEX IF EXISTS idx_household_invitations_invite_code;
DROP INDEX IF EXISTS idx_household_invitations_household_id;
DROP INDEX IF EXISTS idx_household_members_left_at;
DROP INDEX IF EXISTS idx_household_members_user_id;
DROP INDEX IF EXISTS idx_household_members_household_id;
DROP INDEX IF EXISTS idx_households_deleted_at;
DROP INDEX IF EXISTS idx_households_created_by;
DROP INDEX IF EXISTS idx_users_deleted_at;
DROP INDEX IF EXISTS idx_users_phone;
DROP INDEX IF EXISTS idx_users_email;

-- Drop tables in reverse order
DROP TABLE IF EXISTS household_invitations;
DROP TABLE IF EXISTS household_members;
DROP TABLE IF EXISTS user_profiles;
DROP TABLE IF EXISTS households;
DROP TABLE IF EXISTS users;