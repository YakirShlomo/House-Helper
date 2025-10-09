import WidgetKit
import SwiftUI
import Intents

struct Provider: TimelineProvider {
    func placeholder(in context: Context) -> TaskEntry {
        TaskEntry(date: Date(), tasks: sampleTasks)
    }

    func getSnapshot(in context: Context, completion: @escaping (TaskEntry) -> ()) {
        let entry = TaskEntry(date: Date(), tasks: sampleTasks)
        completion(entry)
    }

    func getTimeline(in context: Context, completion: @escaping (Timeline<Entry>) -> ()) {
        var entries: [TaskEntry] = []

        // Generate a timeline consisting of five entries an hour apart, starting from the current date.
        let currentDate = Date()
        for hourOffset in 0 ..< 5 {
            let entryDate = Calendar.current.date(byAdding: .hour, value: hourOffset, to: currentDate)!
            let entry = TaskEntry(date: entryDate, tasks: loadTasks())
            entries.append(entry)
        }

        let timeline = Timeline(entries: entries, policy: .atEnd)
        completion(timeline)
    }
    
    func loadTasks() -> [TaskItem] {
        // TODO: Load from shared UserDefaults or API
        return sampleTasks
    }
}

struct TaskEntry: TimelineEntry {
    let date: Date
    let tasks: [TaskItem]
}

struct TaskItem {
    let id: String
    let title: String
    let isCompleted: Bool
    let dueDate: Date?
}

let sampleTasks = [
    TaskItem(id: "1", title: "Take out trash", isCompleted: false, dueDate: Date()),
    TaskItem(id: "2", title: "Water plants", isCompleted: true, dueDate: nil),
    TaskItem(id: "3", title: "Grocery shopping", isCompleted: false, dueDate: Calendar.current.date(byAdding: .day, value: 1, to: Date()))
]

struct HouseHelperWidgetEntryView : View {
    var entry: Provider.Entry
    @Environment(\.widgetFamily) var family

    var body: some View {
        switch family {
        case .systemSmall:
            SmallTaskWidget(tasks: entry.tasks)
        case .systemMedium:
            MediumTaskWidget(tasks: entry.tasks)
        case .systemLarge:
            LargeTaskWidget(tasks: entry.tasks)
        default:
            SmallTaskWidget(tasks: entry.tasks)
        }
    }
}

struct SmallTaskWidget: View {
    let tasks: [TaskItem]
    
    var body: some View {
        VStack(alignment: .leading, spacing: 4) {
            HStack {
                Image(systemName: "house.fill")
                    .foregroundColor(.blue)
                Text("House Helper")
                    .font(.caption)
                    .fontWeight(.semibold)
                Spacer()
            }
            
            Divider()
            
            let pendingTasks = tasks.filter { !$0.isCompleted }.prefix(2)
            ForEach(Array(pendingTasks.enumerated()), id: \.offset) { index, task in
                HStack {
                    Circle()
                        .fill(Color.gray.opacity(0.3))
                        .frame(width: 8, height: 8)
                    Text(task.title)
                        .font(.caption2)
                        .lineLimit(1)
                    Spacer()
                }
            }
            
            if pendingTasks.isEmpty {
                Text("All done! 🎉")
                    .font(.caption)
                    .foregroundColor(.secondary)
            }
            
            Spacer()
        }
        .padding()
        .background(Color(.systemBackground))
    }
}

struct MediumTaskWidget: View {
    let tasks: [TaskItem]
    
    var body: some View {
        VStack(alignment: .leading, spacing: 8) {
            HStack {
                Image(systemName: "house.fill")
                    .foregroundColor(.blue)
                Text("House Helper")
                    .font(.headline)
                    .fontWeight(.semibold)
                Spacer()
                Text("\(tasks.filter { !$0.isCompleted }.count) pending")
                    .font(.caption)
                    .foregroundColor(.secondary)
            }
            
            Divider()
            
            let pendingTasks = tasks.filter { !$0.isCompleted }.prefix(4)
            ForEach(Array(pendingTasks.enumerated()), id: \.offset) { index, task in
                HStack {
                    Button(intent: CompleteTaskIntent(taskId: task.id)) {
                        Image(systemName: "circle")
                            .foregroundColor(.blue)
                    }
                    .buttonStyle(PlainButtonStyle())
                    
                    VStack(alignment: .leading, spacing: 2) {
                        Text(task.title)
                            .font(.subheadline)
                        if let dueDate = task.dueDate {
                            Text("Due: \(dueDate, style: .date)")
                                .font(.caption2)
                                .foregroundColor(.secondary)
                        }
                    }
                    Spacer()
                }
            }
            
            if pendingTasks.isEmpty {
                HStack {
                    Spacer()
                    VStack {
                        Text("🎉")
                            .font(.title)
                        Text("All tasks completed!")
                            .font(.subheadline)
                            .multilineTextAlignment(.center)
                    }
                    Spacer()
                }
            }
            
            Spacer()
        }
        .padding()
        .background(Color(.systemBackground))
    }
}

struct LargeTaskWidget: View {
    let tasks: [TaskItem]
    
    var body: some View {
        VStack(alignment: .leading, spacing: 12) {
            HStack {
                Image(systemName: "house.fill")
                    .foregroundColor(.blue)
                    .font(.title2)
                VStack(alignment: .leading) {
                    Text("House Helper")
                        .font(.title2)
                        .fontWeight(.bold)
                    Text("Today's Tasks")
                        .font(.subheadline)
                        .foregroundColor(.secondary)
                }
                Spacer()
            }
            
            Divider()
            
            // Pending tasks
            let pendingTasks = tasks.filter { !$0.isCompleted }
            if !pendingTasks.isEmpty {
                Text("Pending (\(pendingTasks.count))")
                    .font(.headline)
                    .foregroundColor(.primary)
                
                ForEach(Array(pendingTasks.prefix(5).enumerated()), id: \.offset) { index, task in
                    HStack {
                        Button(intent: CompleteTaskIntent(taskId: task.id)) {
                            Image(systemName: "circle")
                                .foregroundColor(.blue)
                                .font(.title3)
                        }
                        .buttonStyle(PlainButtonStyle())
                        
                        VStack(alignment: .leading, spacing: 4) {
                            Text(task.title)
                                .font(.body)
                            if let dueDate = task.dueDate {
                                Text("Due: \(dueDate, style: .date)")
                                    .font(.caption)
                                    .foregroundColor(.secondary)
                            }
                        }
                        Spacer()
                    }
                    .padding(.vertical, 2)
                }
            }
            
            // Completed tasks
            let completedTasks = tasks.filter { $0.isCompleted }
            if !completedTasks.isEmpty {
                Text("Completed (\(completedTasks.count))")
                    .font(.headline)
                    .foregroundColor(.green)
                
                ForEach(Array(completedTasks.prefix(3).enumerated()), id: \.offset) { index, task in
                    HStack {
                        Image(systemName: "checkmark.circle.fill")
                            .foregroundColor(.green)
                        Text(task.title)
                            .font(.body)
                            .strikethrough()
                            .foregroundColor(.secondary)
                        Spacer()
                    }
                    .padding(.vertical, 2)
                }
            }
            
            if pendingTasks.isEmpty && completedTasks.isEmpty {
                Spacer()
                HStack {
                    Spacer()
                    VStack {
                        Text("📝")
                            .font(.largeTitle)
                        Text("No tasks yet")
                            .font(.headline)
                        Text("Add your first task in the app")
                            .font(.subheadline)
                            .foregroundColor(.secondary)
                            .multilineTextAlignment(.center)
                    }
                    Spacer()
                }
                Spacer()
            } else {
                Spacer()
            }
        }
        .padding()
        .background(Color(.systemBackground))
    }
}

@main
struct HouseHelperWidget: Widget {
    let kind: String = "HouseHelperWidget"

    var body: some WidgetConfiguration {
        StaticConfiguration(kind: kind, provider: Provider()) { entry in
            HouseHelperWidgetEntryView(entry: entry)
        }
        .configurationDisplayName("House Helper Tasks")
        .description("View your pending household tasks at a glance.")
        .supportedFamilies([.systemSmall, .systemMedium, .systemLarge])
    }
}