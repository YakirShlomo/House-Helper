// ignore: unused_import
import 'package:intl/intl.dart' as intl;
import 'app_localizations.dart';

// ignore_for_file: type=lint

/// The translations for English (`en`).
class AppLocalizationsEn extends AppLocalizations {
  AppLocalizationsEn([String locale = 'en']) : super(locale);

  @override
  String get appTitle => 'House Helper';

  @override
  String get homeTitle => 'Home';

  @override
  String get tasksTitle => 'Tasks';

  @override
  String get shoppingTitle => 'Shopping';

  @override
  String get billsTitle => 'Bills';

  @override
  String get activityTitle => 'Activity';

  @override
  String get settingsTitle => 'Settings';

  @override
  String get addTask => 'Add Task';

  @override
  String get addShoppingItem => 'Add Item';

  @override
  String get addBill => 'Add Bill';

  @override
  String get startTimer => 'Start Timer';

  @override
  String get stopTimer => 'Stop Timer';

  @override
  String get taskCompleted => 'Task Completed';

  @override
  String get laundryTimer => 'Laundry Timer';

  @override
  String get cookingTimer => 'Cooking Timer';

  @override
  String get notifications => 'Notifications';

  @override
  String get profile => 'Profile';

  @override
  String get language => 'Language';

  @override
  String get darkMode => 'Dark Mode';

  @override
  String get logout => 'Logout';

  @override
  String get login => 'Login';

  @override
  String get email => 'Email';

  @override
  String get password => 'Password';

  @override
  String get forgotPassword => 'Forgot Password?';

  @override
  String get signUp => 'Sign Up';

  @override
  String get welcome => 'Welcome to House Helper';

  @override
  String get manageHousehold => 'Manage your household with ease';

  @override
  String get error => 'Error';

  @override
  String get errorDetails => 'Error Details';

  @override
  String get retry => 'Retry';

  @override
  String get networkError =>
      'Network connection failed. Please check your internet connection and try again.';

  @override
  String get authError => 'Authentication required. Please log in to continue.';
}
