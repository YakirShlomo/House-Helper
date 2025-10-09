package app.househelper.mobile

import android.appwidget.AppWidgetManager
import android.appwidget.AppWidgetProvider
import android.content.Context
import android.content.Intent
import android.widget.RemoteViews
import android.app.PendingIntent
import android.content.ComponentName
import androidx.work.WorkManager
import androidx.work.OneTimeWorkRequestBuilder
import androidx.work.Data

class TaskWidgetProvider : AppWidgetProvider() {

    override fun onUpdate(
        context: Context,
        appWidgetManager: AppWidgetManager,
        appWidgetIds: IntArray
    ) {
        // Update all widget instances
        for (appWidgetId in appWidgetIds) {
            updateAppWidget(context, appWidgetManager, appWidgetId)
        }
    }

    override fun onReceive(context: Context, intent: Intent) {
        super.onReceive(context, intent)
        
        when (intent.action) {
            "COMPLETE_TASK" -> {
                val taskId = intent.getStringExtra("task_id") ?: return
                handleCompleteTask(context, taskId)
            }
            "ADD_TASK" -> {
                val taskTitle = intent.getStringExtra("task_title") ?: return
                handleAddTask(context, taskTitle)
            }
            "START_TIMER" -> {
                val timerType = intent.getStringExtra("timer_type") ?: return
                val duration = intent.getIntExtra("duration", 0)
                handleStartTimer(context, timerType, duration)
            }
            "REFRESH_WIDGET" -> {
                val appWidgetManager = AppWidgetManager.getInstance(context)
                val componentName = ComponentName(context, TaskWidgetProvider::class.java)
                val appWidgetIds = appWidgetManager.getAppWidgetIds(componentName)
                onUpdate(context, appWidgetManager, appWidgetIds)
            }
        }
    }

    private fun updateAppWidget(
        context: Context,
        appWidgetManager: AppWidgetManager,
        appWidgetId: Int
    ) {
        val views = RemoteViews(context.packageName, R.layout.task_widget)
        
        // Load tasks from SharedPreferences or local database
        val tasks = loadTasks(context)
        val pendingTasks = tasks.filter { !it.isCompleted }.take(3)
        
        // Update widget title
        views.setTextViewText(R.id.widget_title, "House Helper (${pendingTasks.size} pending)")
        
        // Clear previous task views
        views.removeAllViews(R.id.task_container)
        
        // Add pending tasks
        for ((index, task) in pendingTasks.withIndex()) {
            val taskView = RemoteViews(context.packageName, R.layout.task_item)
            taskView.setTextViewText(R.id.task_title, task.title)
            
            // Set click listener to complete task
            val completeIntent = Intent(context, TaskWidgetProvider::class.java).apply {
                action = "COMPLETE_TASK"
                putExtra("task_id", task.id)
            }
            val completePendingIntent = PendingIntent.getBroadcast(
                context, 
                task.id.hashCode(), 
                completeIntent, 
                PendingIntent.FLAG_UPDATE_CURRENT or PendingIntent.FLAG_IMMUTABLE
            )
            taskView.setOnClickPendingIntent(R.id.task_complete_button, completePendingIntent)
            
            views.addView(R.id.task_container, taskView)
        }
        
        // Add "Add Task" button
        val addTaskIntent = Intent(context, MainActivity::class.java).apply {
            action = "ADD_TASK"
            flags = Intent.FLAG_ACTIVITY_NEW_TASK or Intent.FLAG_ACTIVITY_CLEAR_TOP
        }
        val addTaskPendingIntent = PendingIntent.getActivity(
            context, 
            0, 
            addTaskIntent, 
            PendingIntent.FLAG_UPDATE_CURRENT or PendingIntent.FLAG_IMMUTABLE
        )
        views.setOnClickPendingIntent(R.id.add_task_button, addTaskPendingIntent)
        
        // Add "Start Timer" button
        val startTimerIntent = Intent(context, MainActivity::class.java).apply {
            action = "START_TIMER"
            putExtra("timer_type", "laundry")
            flags = Intent.FLAG_ACTIVITY_NEW_TASK or Intent.FLAG_ACTIVITY_CLEAR_TOP
        }
        val startTimerPendingIntent = PendingIntent.getActivity(
            context, 
            1, 
            startTimerIntent, 
            PendingIntent.FLAG_UPDATE_CURRENT or PendingIntent.FLAG_IMMUTABLE
        )
        views.setOnClickPendingIntent(R.id.start_timer_button, startTimerPendingIntent)
        
        // Update widget
        appWidgetManager.updateAppWidget(appWidgetId, views)
    }

    private fun loadTasks(context: Context): List<TaskItem> {
        val sharedPrefs = context.getSharedPreferences("house_helper_widget", Context.MODE_PRIVATE)
        val tasksJson = sharedPrefs.getString("tasks", null)
        
        // TODO: Parse JSON to TaskItem objects
        // For now, return sample data
        return listOf(
            TaskItem("1", "Take out trash", false),
            TaskItem("2", "Water plants", true),
            TaskItem("3", "Grocery shopping", false)
        )
    }
    
    private fun handleCompleteTask(context: Context, taskId: String) {
        // Mark task as completed in local storage
        val sharedPrefs = context.getSharedPreferences("house_helper_widget", Context.MODE_PRIVATE)
        // TODO: Update task status
        
        // Schedule sync with main app
        scheduleDataSync(context, "complete_task", mapOf("task_id" to taskId))
        
        // Refresh widget
        refreshWidget(context)
    }
    
    private fun handleAddTask(context: Context, taskTitle: String) {
        // Open main app to add task
        val intent = Intent(context, MainActivity::class.java).apply {
            action = "ADD_TASK"
            putExtra("task_title", taskTitle)
            flags = Intent.FLAG_ACTIVITY_NEW_TASK or Intent.FLAG_ACTIVITY_CLEAR_TOP
        }
        context.startActivity(intent)
    }
    
    private fun handleStartTimer(context: Context, timerType: String, duration: Int) {
        // Open main app to start timer
        val intent = Intent(context, MainActivity::class.java).apply {
            action = "START_TIMER"
            putExtra("timer_type", timerType)
            putExtra("duration", duration)
            flags = Intent.FLAG_ACTIVITY_NEW_TASK or Intent.FLAG_ACTIVITY_CLEAR_TOP
        }
        context.startActivity(intent)
    }
    
    private fun scheduleDataSync(context: Context, action: String, data: Map<String, String>) {
        val workData = Data.Builder()
            .putString("action", action)
            .putAll(data)
            .build()
            
        val syncRequest = OneTimeWorkRequestBuilder<DataSyncWorker>()
            .setInputData(workData)
            .build()
            
        WorkManager.getInstance(context).enqueue(syncRequest)
    }
    
    private fun refreshWidget(context: Context) {
        val intent = Intent(context, TaskWidgetProvider::class.java).apply {
            action = "REFRESH_WIDGET"
        }
        context.sendBroadcast(intent)
    }
}

data class TaskItem(
    val id: String,
    val title: String,
    val isCompleted: Boolean,
    val dueDate: Long? = null
)