import Foundation
import Intents
import IntentsUI

// MARK: - Task Management Intents

@available(iOS 12.0, *)
class CompleteTaskIntent: INIntent {
    @NSManaged public var taskId: String?
    
    convenience init(taskId: String) {
        self.init()
        self.taskId = taskId
    }
}

@available(iOS 12.0, *)
class AddTaskIntent: INIntent {
    @NSManaged public var taskTitle: String?
    
    convenience init(taskTitle: String) {
        self.init()
        self.taskTitle = taskTitle
    }
}

@available(iOS 12.0, *)
class StartTimerIntent: INIntent {
    @NSManaged public var timerType: String?
    @NSManaged public var duration: NSNumber?
    
    convenience init(timerType: String, duration: Int) {
        self.init()
        self.timerType = timerType
        self.duration = NSNumber(value: duration)
    }
}

// MARK: - Timer Intents

@available(iOS 14.0, *)
struct StartLaundryTimerIntent: AppIntent {
    static var title: LocalizedStringResource = "Start Laundry Timer"
    static var description = IntentDescription("Start a 45-minute laundry timer")
    
    @Parameter(title: "Task ID")
    var taskId: String?
    
    func perform() async throws -> some IntentResult {
        // Communicate with Flutter app via platform channel or shared UserDefaults
        let userDefaults = UserDefaults(suiteName: "group.app.househelper.mobile")
        userDefaults?.set([
            "action": "start_timer",
            "type": "laundry",
            "duration": 45 * 60,
            "taskId": taskId ?? ""
        ], forKey: "intent_action")
        
        // Open app to specific timer
        await OpenUrlIntent(url: URL(string: "househelper://timer/laundry")!).perform()
        
        return .result(dialog: "Laundry timer started for 45 minutes")
    }
    
    static var openAppWhenRun: Bool = true
}

@available(iOS 14.0, *)
struct AddShoppingItemIntent: AppIntent {
    static var title: LocalizedStringResource = "Add Shopping Item"
    static var description = IntentDescription("Add an item to your shopping list")
    
    @Parameter(title: "Item Name")
    var itemName: String
    
    func perform() async throws -> some IntentResult {
        let userDefaults = UserDefaults(suiteName: "group.app.househelper.mobile")
        userDefaults?.set([
            "action": "add_shopping_item",
            "name": itemName
        ], forKey: "intent_action")
        
        await OpenUrlIntent(url: URL(string: "househelper://shopping/add?name=\(itemName.addingPercentEncoding(withAllowedCharacters: .urlQueryAllowed) ?? "")")!).perform()
        
        return .result(dialog: "Added \(itemName) to your shopping list")
    }
    
    static var openAppWhenRun: Bool = true
}

@available(iOS 14.0, *)
struct MarkTaskDoneIntent: AppIntent {
    static var title: LocalizedStringResource = "Mark Task Done"
    static var description = IntentDescription("Mark a household task as completed")
    
    @Parameter(title: "Task ID")
    var taskId: String
    
    func perform() async throws -> some IntentResult {
        let userDefaults = UserDefaults(suiteName: "group.app.househelper.mobile")
        userDefaults?.set([
            "action": "complete_task",
            "taskId": taskId
        ], forKey: "intent_action")
        
        await OpenUrlIntent(url: URL(string: "househelper://tasks/\(taskId)/complete")!).perform()
        
        return .result(dialog: "Task marked as completed")
    }
    
    static var openAppWhenRun: Bool = true
}

// MARK: - Intent Handler

@available(iOS 12.0, *)
class IntentHandler: INExtension {
    
    override func handler(for intent: INIntent) -> Any {
        switch intent {
        case is CompleteTaskIntent:
            return CompleteTaskIntentHandler()
        case is AddTaskIntent:
            return AddTaskIntentHandler()
        case is StartTimerIntent:
            return StartTimerIntentHandler()
        default:
            fatalError("Unhandled intent type: \(intent)")
        }
    }
}

@available(iOS 12.0, *)
class CompleteTaskIntentHandler: NSObject, CompleteTaskIntentHandling {
    func handle(intent: CompleteTaskIntent, completion: @escaping (CompleteTaskIntentResponse) -> Void) {
        guard let taskId = intent.taskId else {
            completion(CompleteTaskIntentResponse(code: .failure, userActivity: nil))
            return
        }
        
        // Store action for Flutter app to pick up
        let userDefaults = UserDefaults(suiteName: "group.app.househelper.mobile")
        userDefaults?.set([
            "action": "complete_task",
            "taskId": taskId,
            "timestamp": Date().timeIntervalSince1970
        ], forKey: "intent_action")
        
        completion(CompleteTaskIntentResponse(code: .success, userActivity: nil))
    }
}

@available(iOS 12.0, *)
class AddTaskIntentHandler: NSObject, AddTaskIntentHandling {
    func handle(intent: AddTaskIntent, completion: @escaping (AddTaskIntentResponse) -> Void) {
        guard let title = intent.taskTitle, !title.isEmpty else {
            completion(AddTaskIntentResponse(code: .failure, userActivity: nil))
            return
        }
        
        let userDefaults = UserDefaults(suiteName: "group.app.househelper.mobile")
        userDefaults?.set([
            "action": "add_task",
            "title": title,
            "timestamp": Date().timeIntervalSince1970
        ], forKey: "intent_action")
        
        completion(AddTaskIntentResponse(code: .success, userActivity: nil))
    }
}

@available(iOS 12.0, *)
class StartTimerIntentHandler: NSObject, StartTimerIntentHandling {
    func handle(intent: StartTimerIntent, completion: @escaping (StartTimerIntentResponse) -> Void) {
        guard let timerType = intent.timerType,
              let duration = intent.duration else {
            completion(StartTimerIntentResponse(code: .failure, userActivity: nil))
            return
        }
        
        let userDefaults = UserDefaults(suiteName: "group.app.househelper.mobile")
        userDefaults?.set([
            "action": "start_timer",
            "type": timerType,
            "duration": duration.intValue,
            "timestamp": Date().timeIntervalSince1970
        ], forKey: "intent_action")
        
        completion(StartTimerIntentResponse(code: .success, userActivity: nil))
    }
}

// MARK: - Protocol Extensions

@available(iOS 12.0, *)
protocol CompleteTaskIntentHandling {
    func handle(intent: CompleteTaskIntent, completion: @escaping (CompleteTaskIntentResponse) -> Void)
}

@available(iOS 12.0, *)
protocol AddTaskIntentHandling {
    func handle(intent: AddTaskIntent, completion: @escaping (AddTaskIntentResponse) -> Void)
}

@available(iOS 12.0, *)
protocol StartTimerIntentHandling {
    func handle(intent: StartTimerIntent, completion: @escaping (StartTimerIntentResponse) -> Void)
}

// MARK: - Intent Responses

@available(iOS 12.0, *)
class CompleteTaskIntentResponse: INIntentResponse {
    convenience init(code: CompleteTaskIntentResponseCode, userActivity: NSUserActivity?) {
        self.init()
        self.code = code
        self.userActivity = userActivity
    }
    
    var code: CompleteTaskIntentResponseCode = .unspecified
}

@available(iOS 12.0, *)
enum CompleteTaskIntentResponseCode: Int {
    case unspecified = 0
    case ready = 1
    case continueInApp = 2
    case inProgress = 3
    case success = 4
    case failure = 5
    case failureRequiringAppLaunch = 6
}

@available(iOS 12.0, *)
class AddTaskIntentResponse: INIntentResponse {
    convenience init(code: AddTaskIntentResponseCode, userActivity: NSUserActivity?) {
        self.init()
        self.code = code
        self.userActivity = userActivity
    }
    
    var code: AddTaskIntentResponseCode = .unspecified
}

@available(iOS 12.0, *)
enum AddTaskIntentResponseCode: Int {
    case unspecified = 0
    case ready = 1
    case continueInApp = 2
    case inProgress = 3
    case success = 4
    case failure = 5
    case failureRequiringAppLaunch = 6
}

@available(iOS 12.0, *)
class StartTimerIntentResponse: INIntentResponse {
    convenience init(code: StartTimerIntentResponseCode, userActivity: NSUserActivity?) {
        self.init()
        self.code = code
        self.userActivity = userActivity
    }
    
    var code: StartTimerIntentResponseCode = .unspecified
}

@available(iOS 12.0, *)
enum StartTimerIntentResponseCode: Int {
    case unspecified = 0
    case ready = 1
    case continueInApp = 2
    case inProgress = 3
    case success = 4
    case failure = 5
    case failureRequiringAppLaunch = 6
}