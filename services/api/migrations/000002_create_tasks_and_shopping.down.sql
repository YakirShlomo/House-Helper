-- Drop triggers
DROP TRIGGER IF EXISTS update_products_updated_at ON products;
DROP TRIGGER IF EXISTS update_shopping_list_shares_updated_at ON shopping_list_shares;
DROP TRIGGER IF EXISTS update_shopping_items_updated_at ON shopping_items;
DROP TRIGGER IF EXISTS update_shopping_lists_updated_at ON shopping_lists;
DROP TRIGGER IF EXISTS update_tasks_updated_at ON tasks;

-- Drop indexes
DROP INDEX IF EXISTS idx_products_brand;
DROP INDEX IF EXISTS idx_products_barcode;
DROP INDEX IF EXISTS idx_products_category;
DROP INDEX IF EXISTS idx_products_name;
DROP INDEX IF EXISTS idx_shopping_list_shares_user_id;
DROP INDEX IF EXISTS idx_shopping_list_shares_list_id;
DROP INDEX IF EXISTS idx_shopping_items_deleted_at;
DROP INDEX IF EXISTS idx_shopping_items_barcode;
DROP INDEX IF EXISTS idx_shopping_items_category;
DROP INDEX IF EXISTS idx_shopping_items_is_purchased;
DROP INDEX IF EXISTS idx_shopping_items_purchased_by;
DROP INDEX IF EXISTS idx_shopping_items_added_by;
DROP INDEX IF EXISTS idx_shopping_items_list_id;
DROP INDEX IF EXISTS idx_shopping_lists_deleted_at;
DROP INDEX IF EXISTS idx_shopping_lists_created_by;
DROP INDEX IF EXISTS idx_shopping_lists_household_id;
DROP INDEX IF EXISTS idx_tasks_deleted_at;
DROP INDEX IF EXISTS idx_tasks_due_date;
DROP INDEX IF EXISTS idx_tasks_category;
DROP INDEX IF EXISTS idx_tasks_priority;
DROP INDEX IF EXISTS idx_tasks_status;
DROP INDEX IF EXISTS idx_tasks_created_by;
DROP INDEX IF EXISTS idx_tasks_assigned_to;
DROP INDEX IF EXISTS idx_tasks_household_id;

-- Drop tables in reverse order
DROP TABLE IF EXISTS products;
DROP TABLE IF EXISTS shopping_list_shares;
DROP TABLE IF EXISTS shopping_items;
DROP TABLE IF EXISTS shopping_lists;
DROP TABLE IF EXISTS tasks;