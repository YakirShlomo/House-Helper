package app.househelper.mobile

import android.content.Context
import android.content.Intent
import android.content.pm.ShortcutInfo
import android.content.pm.ShortcutManager
import android.graphics.drawable.Icon
import android.os.Build
import androidx.annotation.RequiresApi

object ShortcutManager {
    
    @RequiresApi(Build.VERSION_CODES.N_MR1)
    fun createStaticShortcuts(context: Context) {
        val shortcutManager = context.getSystemService(ShortcutManager::class.java)
        
        val shortcuts = listOf(
            createAddTaskShortcut(context),
            createStartTimerShortcut(context),
            createViewTasksShortcut(context),
            createAddShoppingItemShortcut(context)
        )
        
        shortcutManager?.dynamicShortcuts = shortcuts
    }
    
    @RequiresApi(Build.VERSION_CODES.N_MR1)
    private fun createAddTaskShortcut(context: Context): ShortcutInfo {
        val intent = Intent(context, MainActivity::class.java).apply {
            action = "ADD_TASK"
            flags = Intent.FLAG_ACTIVITY_NEW_TASK or Intent.FLAG_ACTIVITY_CLEAR_TASK
        }
        
        return ShortcutInfo.Builder(context, "add_task")
            .setShortLabel("Add Task")
            .setLongLabel("Add New Task")
            .setIcon(Icon.createWithResource(context, R.drawable.ic_add_task))
            .setIntent(intent)
            .build()
    }
    
    @RequiresApi(Build.VERSION_CODES.N_MR1)
    private fun createStartTimerShortcut(context: Context): ShortcutInfo {
        val intent = Intent(context, MainActivity::class.java).apply {
            action = "START_TIMER"
            putExtra("timer_type", "laundry")
            putExtra("duration", 45 * 60) // 45 minutes
            flags = Intent.FLAG_ACTIVITY_NEW_TASK or Intent.FLAG_ACTIVITY_CLEAR_TASK
        }
        
        return ShortcutInfo.Builder(context, "start_laundry_timer")
            .setShortLabel("Laundry Timer")
            .setLongLabel("Start Laundry Timer")
            .setIcon(Icon.createWithResource(context, R.drawable.ic_timer))
            .setIntent(intent)
            .build()
    }
    
    @RequiresApi(Build.VERSION_CODES.N_MR1)
    private fun createViewTasksShortcut(context: Context): ShortcutInfo {
        val intent = Intent(context, MainActivity::class.java).apply {
            action = "VIEW_TASKS"
            flags = Intent.FLAG_ACTIVITY_NEW_TASK or Intent.FLAG_ACTIVITY_CLEAR_TASK
        }
        
        return ShortcutInfo.Builder(context, "view_tasks")
            .setShortLabel("Tasks")
            .setLongLabel("View All Tasks")
            .setIcon(Icon.createWithResource(context, R.drawable.ic_tasks))
            .setIntent(intent)
            .build()
    }
    
    @RequiresApi(Build.VERSION_CODES.N_MR1)
    private fun createAddShoppingItemShortcut(context: Context): ShortcutInfo {
        val intent = Intent(context, MainActivity::class.java).apply {
            action = "ADD_SHOPPING_ITEM"
            flags = Intent.FLAG_ACTIVITY_NEW_TASK or Intent.FLAG_ACTIVITY_CLEAR_TASK
        }
        
        return ShortcutInfo.Builder(context, "add_shopping_item")
            .setShortLabel("Add Item")
            .setLongLabel("Add Shopping Item")
            .setIcon(Icon.createWithResource(context, R.drawable.ic_shopping))
            .setIntent(intent)
            .build()
    }
    
    @RequiresApi(Build.VERSION_CODES.N_MR1)
    fun updateDynamicShortcuts(context: Context, tasks: List<TaskItem>) {
        val shortcutManager = context.getSystemService(ShortcutManager::class.java)
        
        // Create shortcuts for recent tasks
        val taskShortcuts = tasks.filter { !it.isCompleted }
            .take(3)
            .mapIndexed { index, task ->
                val intent = Intent(context, MainActivity::class.java).apply {
                    action = "COMPLETE_TASK"
                    putExtra("task_id", task.id)
                    flags = Intent.FLAG_ACTIVITY_NEW_TASK or Intent.FLAG_ACTIVITY_CLEAR_TASK
                }
                
                ShortcutInfo.Builder(context, "task_${task.id}")
                    .setShortLabel("✓ ${task.title.take(10)}")
                    .setLongLabel("Complete: ${task.title}")
                    .setIcon(Icon.createWithResource(context, R.drawable.ic_check))
                    .setIntent(intent)
                    .setRank(index)
                    .build()
            }
        
        // Combine with static shortcuts
        val allShortcuts = mutableListOf<ShortcutInfo>().apply {
            addAll(taskShortcuts)
            if (size < 4) {
                add(createAddTaskShortcut(context))
            }
            if (size < 4) {
                add(createStartTimerShortcut(context))
            }
        }
        
        shortcutManager?.dynamicShortcuts = allShortcuts.take(4)
    }
    
    @RequiresApi(Build.VERSION_CODES.N_MR1)
    fun reportShortcutUsed(context: Context, shortcutId: String) {
        val shortcutManager = context.getSystemService(ShortcutManager::class.java)
        shortcutManager?.reportShortcutUsed(shortcutId)
    }
}