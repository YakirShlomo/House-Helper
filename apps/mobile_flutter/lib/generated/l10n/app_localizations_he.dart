// ignore: unused_import
import 'package:intl/intl.dart' as intl;
import 'app_localizations.dart';

// ignore_for_file: type=lint

/// The translations for Hebrew (`he`).
class AppLocalizationsHe extends AppLocalizations {
  AppLocalizationsHe([String locale = 'he']) : super(locale);

  @override
  String get appTitle => 'עוזר הבית';

  @override
  String get homeTitle => 'בית';

  @override
  String get tasksTitle => 'משימות';

  @override
  String get shoppingTitle => 'קניות';

  @override
  String get billsTitle => 'חשבונות';

  @override
  String get activityTitle => 'פעילות';

  @override
  String get settingsTitle => 'הגדרות';

  @override
  String get addTask => 'הוסף משימה';

  @override
  String get addShoppingItem => 'הוסף פריט';

  @override
  String get addBill => 'הוסף חשבון';

  @override
  String get startTimer => 'התחל שעון עצר';

  @override
  String get stopTimer => 'עצור שעון עצר';

  @override
  String get taskCompleted => 'משימה הושלמה';

  @override
  String get laundryTimer => 'שעון עצר לכביסה';

  @override
  String get cookingTimer => 'שעון עצר לבישול';

  @override
  String get notifications => 'התראות';

  @override
  String get profile => 'פרופיל';

  @override
  String get language => 'שפה';

  @override
  String get darkMode => 'מצב כהה';

  @override
  String get logout => 'התנתק';

  @override
  String get login => 'התחבר';

  @override
  String get email => 'אימייל';

  @override
  String get password => 'סיסמה';

  @override
  String get forgotPassword => 'שכחת סיסמה?';

  @override
  String get signUp => 'הרשם';

  @override
  String get welcome => 'ברוכים הבאים לעוזר הבית';

  @override
  String get manageHousehold => 'נהלו את הבית שלכם בקלות';

  @override
  String get error => 'שגיאה';

  @override
  String get errorDetails => 'פרטי השגיאה';

  @override
  String get retry => 'נסה שוב';

  @override
  String get networkError =>
      'החיבור לאינטרנט נכשל. אנא בדוק את החיבור שלך ונסה שוב.';

  @override
  String get authError => 'נדרשת הזדהות. אנא התחבר כדי להמשיך.';
}
