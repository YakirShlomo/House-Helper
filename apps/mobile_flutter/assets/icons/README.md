# App Icons

## Required Icons

To generate app icons for all platforms, you need:

1. **app_icon.png** - Main app icon (1024x1024px, PNG with transparent background)
2. **app_icon_foreground.png** - Foreground layer for Android adaptive icon (1024x1024px, PNG with transparent background)

## How to Generate Icons

1. Place your icon files in this directory (`assets/icons/`)
2. Run: `flutter pub get`
3. Run: `flutter pub run flutter_launcher_icons`

## Icon Design Guidelines

### Main Icon (app_icon.png)
- Size: 1024x1024 pixels
- Format: PNG with transparency
- Content: Should work well at small sizes
- Background: Transparent or with your brand color (#6750A4)
- Safe zone: Keep important content within the center 60% of the canvas

### Adaptive Icon Foreground (app_icon_foreground.png)
- Size: 1024x1024 pixels  
- Format: PNG with transparency
- Content: Your logo/icon without background
- Safe zone: Keep all content within 66% of the canvas diameter (centered circle)
- Padding: Leave 25% padding on all sides

## Quick Icon Creation

If you don't have an icon yet, you can:

1. Use a simple design tool like Canva or Figma
2. Create a 1024x1024 canvas
3. Add your app name/logo/symbol
4. Use the brand color: #6750A4 (Purple)
5. Export as PNG with transparency

## Temporary Solution

For testing purposes, you can use a simple colored square:
- Background: #6750A4
- Icon: House emoji 🏠 or "HH" text in white
- This is good enough for development and testing

## After Icons Are Generated

The `flutter_launcher_icons` plugin will automatically generate:
- Android: `mipmap-*` folders with different sizes
- iOS: `Assets.xcassets/AppIcon.appiconset`
- Web: `icons/Icon-*.png`
- Windows: `windows/runner/resources/app_icon.ico`
- macOS: `macos/Runner/Assets.xcassets/AppIcon.appiconset`
