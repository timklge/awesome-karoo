# awesome-karoo

A community-driven list of known extensions and resources for the Karoo cycling computer.

## Table of Contents

- [Introduction](#introduction)
- [Extensions](#extensions)
  - [VinHKE](#vinhke)
  - [karoo-kpower](#karoo-kpower)
  - [karoo-kremote](#karoo-kremote)
  - [karoo-kdoubletype](#karoo-kdoubletype)
  - [karoo-reminder](#karoo-reminder)
  - [karoo-powerbar](#karoo-powerbar)
  - [karoo-headwind](#karoo-headwind)
- [Resources](#resources)
  - [karoo-ext](#karoo-ext)
- [Contributing](#contributing)
- [License](#license)

## Introduction

Recently, hammerhead has published a new SDK that enables the development of extensions for Karoo cycling computers. This repository aims to compile a comprehensive list of extensions.

## Installation of extensions on your device

If you are using a Karoo 3, you can use [Hammerhead's sideloading procedure](https://support.hammerhead.io/hc/en-us/articles/31576497036827-Companion-App-Sideloading) to install extensions.

1. Open the releases page of your chosen extension on your phone's browser, long-press the link to the APK (usually `app-release.apk`) and share it with the Hammerhead Companion app.
2. Your karoo should show an info screen about the app now. Press "Install".
3. Open the app from the main menu.

If you are using a Karoo 2, you can use manual sideloading:

1. Download the apk from the releases page of your chosen extension
2. Set up your Karoo for sideloading. DC Rainmaker has a great [step-by-step guide](https://www.dcrainmaker.com/2021/02/how-to-sideload-android-apps-on-your-hammerhead-karoo-1-karoo-2.html).
3. Install the app by running `adb install FILENAME.apk`.
4. Open the app from the main menu.

To update extensions after you have installed them, long-tap the app icon on the main menu and select "Update" (works on both Karoo 2 and Karoo 3).

## Extensions

### VinHKE
- **Extension Name:** [Vin-HkE](https://github.com/maduwatas/Vin-HkE)
  - Description: Vin HkE is an extension implementation for Hammerhead Karroo based on Vin's CORE and the new Hammerhead SDK for its GPS devices. Tested in Hammerhead K2 and K3 devices. Vin is an Android training and planning app for cycling with more than 4 years of life and a 5-star rating on Google Play. 
  - License: Closed Source, Freeware
  - Features:
    - Workout player
    - Virtual pace partner
    - Colored data fields
    - Graphical data fields for power, heart rate
    - Custom alerts based on time, pulse, power, calories burned
    - Pit board

### karoo-kpower
- **Extension Name:** [karoo-kpower](https://github.com/lockevod/Karoo-KPower)
  - Description: Adds a virtual power meter device
  - License: Open Source, Apache 2
  - Features:
    - Virtual power meter that estimates your wattage based on rolling resistance, FTP, headwind, drag coefficient and more

### karoo-kremote
- **Extension Name:** [karoo-kremote](https://github.com/lockevod/Karoo-KRemote)
  - Description: KRemote allows to use a remote with Karoo and perform some actions with it (swipe screens, etc)
  - License: Open Source, Apache 2
  - Features:
    - Use ANT+ remotes with the Karoo to swipe screens or trigger a back button press (map zoom etc. depending on current view)
    
### karoo-kdoubletype
- **Extension Name:** [karoo-KDoubleType](https://github.com/lockevod/Karoo-KDoubleType)
  - Description: This extension for Karoo allows to use custom fields with double types in the same field(HR in right, Power in left, etc)
  - License: Open Source, Apache 2
  - Features:
    - Show fields with two types of data in the same field.

### karoo-reminder
- **Extension Name:** [karoo-reminder](https://github.com/timklge/karoo-reminder)
  - Description: Shows in-ride alerts based on custom triggers
  - License: Open Source, Apache 2
  - Features:
    - Set up reminders that are shown during riding
    - Reminders can be set to activate after a specific time interval, distance traveled, or when a sensor value is outside a defined range (e.g., heart rate exceeds zone 2)
    - Reminders can be set to turn on the screen, play a beep tone and sound over bluetooth, show a custom text and auto-dismiss after a specified duration

### karoo-powerbar
- **Extension Name:** [karoo-powerbar](https://github.com/timklge/karoo-powerbar)
  - Description: Shows an overlay power bar at the edge of the screen. Created as a replacement for the LED bars on wahoo cycling computers.
  - License: Open Source, Apache 2
  - Features:
    - Show colored bars at the top or bottom of the screen that are filled depending on current power output or heart rate
    - Bars are always displayed on top of other apps (i. e. the ride app)

### karoo-headwind
- **Extension Name:** [karoo-headwind](https://github.com/timklge/karoo-headwind)
  - Description: Adds a visual data field showing the wind direction and speed relative to your riding path, as well as an hourly weather forecast.
  - License: Open Source, Apache 2
  - Features:
    - Show current headwind speed and direction in a datafield
    - Periodically fetches weather data from [open-meteo.com](https://open-meteo.com)
    - Show current weather conditions and forecasted weather conditions for the next hours in a datafield (precipitation, wind speed, temperature, cloud cover...)
    - Additional data fields for surface pressure, cloud cover etc

## Resources

### karoo-ext

The new SDK to develop extensions is maintained by Hammerhead at https://github.com/hammerheadnav/karoo-ext.

## Contributing

We welcome contributions to this repository! If you have developed an extension or resource that you believe would be beneficial to others, please follow these steps:

1. **Fork** the repository on GitHub.
2. **Create** a new branch for your contribution (`git checkout -b feature/your-extension`).
3. **Add** your extension details in the appropriate section of this README file.
4. **Commit** your changes (`git commit -m 'Add your extension'`).
5. **Push** to the branch (`git push origin feature/your-extension`).
6. **Submit** a pull request.

Please ensure that your contribution includes:
- A clear description of the extension or resource.
- License information.
- Features and benefits of the extension.

## License

This repository is licensed under the Apache 2 License. See [LICENSE](LICENSE) for more details.
