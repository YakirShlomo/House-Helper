package app.househelper.mobile

import android.content.Context
import androidx.work.Worker
import androidx.work.WorkerParameters
import android.content.SharedPreferences
import androidx.work.ListenableWorker
import java.util.concurrent.TimeUnit

class DataSyncWorker(
    context: Context,
    workerParams: WorkerParameters
) : Worker(context, workerParams) {

    override fun doWork(): Result {
        return try {
            val action = inputData.getString("action") ?: return Result.failure()
            
            when (action) {
                "complete_task" -> {
                    val taskId = inputData.getString("task_id") ?: return Result.failure()
                    syncCompleteTask(taskId)
                }
                "add_task" -> {
                    val title = inputData.getString("task_title") ?: return Result.failure()
                    syncAddTask(title)
                }
                "start_timer" -> {
                    val timerType = inputData.getString("timer_type") ?: return Result.failure()
                    val duration = inputData.getInt("duration", 0)
                    syncStartTimer(timerType, duration)
                }
                "sync_tasks" -> {
                    syncTasks()
                }
            }
            
            Result.success()
        } catch (e: Exception) {
            Result.retry()
        }
    }
    
    private fun syncCompleteTask(taskId: String) {
        // Store action for Flutter app to process
        val prefs = applicationContext.getSharedPreferences("house_helper_sync", Context.MODE_PRIVATE)
        val actions = prefs.getStringSet("pending_actions", mutableSetOf()) ?: mutableSetOf()
        
        actions.add("complete_task:$taskId:${System.currentTimeMillis()}")
        
        prefs.edit()
            .putStringSet("pending_actions", actions)
            .apply()
    }
    
    private fun syncAddTask(title: String) {
        val prefs = applicationContext.getSharedPreferences("house_helper_sync", Context.MODE_PRIVATE)
        val actions = prefs.getStringSet("pending_actions", mutableSetOf()) ?: mutableSetOf()
        
        actions.add("add_task:$title:${System.currentTimeMillis()}")
        
        prefs.edit()
            .putStringSet("pending_actions", actions)
            .apply()
    }
    
    private fun syncStartTimer(timerType: String, duration: Int) {
        val prefs = applicationContext.getSharedPreferences("house_helper_sync", Context.MODE_PRIVATE)
        val actions = prefs.getStringSet("pending_actions", mutableSetOf()) ?: mutableSetOf()
        
        actions.add("start_timer:$timerType:$duration:${System.currentTimeMillis()}")
        
        prefs.edit()
            .putStringSet("pending_actions", actions)
            .apply()
    }
    
    private fun syncTasks() {
        // Fetch latest tasks from main app's shared storage
        val mainAppPrefs = applicationContext.getSharedPreferences("flutter.house_helper", Context.MODE_PRIVATE)
        val tasksJson = mainAppPrefs.getString("tasks", null)
        
        if (tasksJson != null) {
            // Update widget's local cache
            val widgetPrefs = applicationContext.getSharedPreferences("house_helper_widget", Context.MODE_PRIVATE)
            widgetPrefs.edit()
                .putString("tasks", tasksJson)
                .putLong("last_sync", System.currentTimeMillis())
                .apply()
                
            // Trigger widget refresh
            refreshWidget()
        }
    }
    
    private fun refreshWidget() {
        val intent = android.content.Intent(applicationContext, TaskWidgetProvider::class.java).apply {
            action = "REFRESH_WIDGET"
        }
        applicationContext.sendBroadcast(intent)
    }
}

// Periodic sync worker
class PeriodicSyncWorker(
    context: Context,
    workerParams: WorkerParameters
) : Worker(context, workerParams) {

    override fun doWork(): Result {
        return try {
            // Sync tasks from main app
            val syncWorker = DataSyncWorker(applicationContext, workerParams)
            syncWorker.doWork()
            
            Result.success()
        } catch (e: Exception) {
            Result.retry()
        }
    }
}

// Utility class for scheduling periodic syncs
object WidgetSyncScheduler {
    
    fun schedulePeriodicSync(context: Context) {
        val workManager = androidx.work.WorkManager.getInstance(context)
        
        val syncRequest = androidx.work.PeriodicWorkRequestBuilder<PeriodicSyncWorker>(
            15, TimeUnit.MINUTES
        ).build()
        
        workManager.enqueueUniquePeriodicWork(
            "widget_sync",
            androidx.work.ExistingPeriodicWorkPolicy.KEEP,
            syncRequest
        )
    }
    
    fun scheduleSyncNow(context: Context) {
        val workManager = androidx.work.WorkManager.getInstance(context)
        
        val syncData = androidx.work.Data.Builder()
            .putString("action", "sync_tasks")
            .build()
            
        val syncRequest = androidx.work.OneTimeWorkRequestBuilder<DataSyncWorker>()
            .setInputData(syncData)
            .build()
            
        workManager.enqueue(syncRequest)
    }
}