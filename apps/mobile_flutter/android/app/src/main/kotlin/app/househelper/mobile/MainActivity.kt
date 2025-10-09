package app.househelper.mobile

import io.flutter.embedding.android.FlutterActivity
import io.flutter.embedding.engine.FlutterEngine
import io.flutter.plugin.common.MethodChannel
import android.content.Intent
import android.os.Build
import androidx.annotation.RequiresApi

class MainActivity: FlutterActivity() {
    private val CHANNEL = "app.househelper.mobile/native"
    private lateinit var methodChannel: MethodChannel

    override fun configureFlutterEngine(flutterEngine: FlutterEngine) {
        super.configureFlutterEngine(flutterEngine)
        
        methodChannel = MethodChannel(flutterEngine.dartExecutor.binaryMessenger, CHANNEL)
        methodChannel.setMethodCallHandler { call, result ->
            when (call.method) {
                "updateWidgets" -> {
                    updateWidgets()
                    result.success(true)
                }
                "createShortcuts" -> {
                    if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.N_MR1) {
                        createShortcuts()
                        result.success(true)
                    } else {
                        result.error("UNSUPPORTED", "Shortcuts require Android 7.1+", null)
                    }
                }
                "updateDynamicShortcuts" -> {
                    if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.N_MR1) {
                        val tasks = call.argument<List<Map<String, Any>>>("tasks")
                        updateDynamicShortcuts(tasks)
                        result.success(true)
                    } else {
                        result.error("UNSUPPORTED", "Dynamic shortcuts require Android 7.1+", null)
                    }
                }
                "scheduleWidgetSync" -> {
                    scheduleWidgetSync()
                    result.success(true)
                }
                "getPendingActions" -> {
                    val actions = getPendingActions()
                    result.success(actions)
                }
                "clearPendingActions" -> {
                    clearPendingActions()
                    result.success(true)
                }
                else -> {
                    result.notImplemented()
                }
            }
        }
        
        // Schedule periodic widget updates
        WidgetSyncScheduler.schedulePeriodicSync(this)
    }

    override fun onNewIntent(intent: Intent) {
        super.onNewIntent(intent)
        handleIntent(intent)
    }

    override fun onResume() {
        super.onResume()
        handleIntent(intent)
    }

    private fun handleIntent(intent: Intent) {
        when (intent.action) {
            "ADD_TASK" -> {
                val title = intent.getStringExtra("task_title")
                methodChannel.invokeMethod("handleIntent", mapOf(
                    "action" to "add_task",
                    "title" to title
                ))
            }
            "START_TIMER" -> {
                val timerType = intent.getStringExtra("timer_type")
                val duration = intent.getIntExtra("duration", 0)
                methodChannel.invokeMethod("handleIntent", mapOf(
                    "action" to "start_timer",
                    "timer_type" to timerType,
                    "duration" to duration
                ))
            }
            "COMPLETE_TASK" -> {
                val taskId = intent.getStringExtra("task_id")
                methodChannel.invokeMethod("handleIntent", mapOf(
                    "action" to "complete_task",
                    "task_id" to taskId
                ))
            }
            "VIEW_TASKS" -> {
                methodChannel.invokeMethod("handleIntent", mapOf(
                    "action" to "view_tasks"
                ))
            }
            "ADD_SHOPPING_ITEM" -> {
                val itemName = intent.getStringExtra("item_name")
                methodChannel.invokeMethod("handleIntent", mapOf(
                    "action" to "add_shopping_item",
                    "item_name" to itemName
                ))
            }
        }
    }

    private fun updateWidgets() {
        val intent = Intent(this, TaskWidgetProvider::class.java).apply {
            action = "REFRESH_WIDGET"
        }
        sendBroadcast(intent)
    }

    @RequiresApi(Build.VERSION_CODES.N_MR1)
    private fun createShortcuts() {
        ShortcutManager.createStaticShortcuts(this)
    }

    @RequiresApi(Build.VERSION_CODES.N_MR1)
    private fun updateDynamicShortcuts(tasksData: List<Map<String, Any>>?) {
        if (tasksData == null) return
        
        val tasks = tasksData.map { taskMap ->
            TaskItem(
                id = taskMap["id"] as String,
                title = taskMap["title"] as String,
                isCompleted = taskMap["isCompleted"] as Boolean
            )
        }
        
        ShortcutManager.updateDynamicShortcuts(this, tasks)
    }

    private fun scheduleWidgetSync() {
        WidgetSyncScheduler.scheduleSyncNow(this)
    }

    private fun getPendingActions(): List<String> {
        val prefs = getSharedPreferences("house_helper_sync", MODE_PRIVATE)
        return prefs.getStringSet("pending_actions", emptySet())?.toList() ?: emptyList()
    }

    private fun clearPendingActions() {
        val prefs = getSharedPreferences("house_helper_sync", MODE_PRIVATE)
        prefs.edit().remove("pending_actions").apply()
    }
}