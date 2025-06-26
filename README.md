# awesome-karoo

A community-driven list of known extensions and resources for the Karoo cycling computer.

## Table of Contents

- [Introduction](#introduction)
- [Extensions](#extensions)
  - [Ki2](#ki2)
  - [VinHKE](#vinhke)
  - [kxradar](#kxradar)
  - [karoo-kpower](#karoo-kpower)
  - [karoo-kremote](#karoo-kremote)
  - [karoo-kdoubletype](#karoo-kdoubletype)
  - [karoo-kactions](#karoo-kactions)
  - [karoo-reminder](#karoo-reminder)
  - [karoo-powerbar](#karoo-powerbar)
  - [karoo-headwind](#karoo-headwind)
  - [karoo-tilehunting](#karoo-tilehunting)
  - [karoo-routegraph](#karoo-routegraph)
  - [Waypoints](#waypoints)
  - [karoo-notepad](#karoo-notepad)
  - [karoo-spintunes](#karoo-spintunes)
- [Libraries](#libraries)
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

### Ki2
- **Extension Name:** [Ki2](https://github.com/valterc/ki2)
  - Description: Ki2 is a companion app for electronic shifting groupsets produced by a Japanese company.
  - License: Open Source, MIT
  - Features:
    - Pair and connect to electronic shifting groupsets
    - Configure actions for hood 
    - Change shifting mode
    - Setup Karoo ride profiles with shifting data elements    
    - While in a ride, control Karoo from the shifting buttons
    - Overlay with shifting information while in Ride
    - Receive notifications when shifting battery is low
    - Audio alerts before upcoming synchro shift or when reaching shifting limits

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

### kxradar
- **Extension Name:** [kxradar](https://github.com/itxsvv/kxradar)
  - Description: Hammerhead Karoo extension which allows to configure radar alerts
  - License: Open Source, Apache 2
  - Features:
    - Replace the radar alert sound with a custom beep

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
    - Show fields with two types of data in the same field
   
### karoo-kactions
- **Extension Name:** [karoo-KActions](https://github.com/lockevod/Karoo-KActions)
  - Description: KActions allows to execute webhooks and send automated messages (whatsapp or sms with live tracking) when you start, stop, your activity in a karoo GPS 
  - License: Open Source, MIT
  - Features
    - Send automated notifications when you start or finish a bike ride. Now compatible with multiple messaging providers
    - Execute automated actions (webhook)
    - Execute custom custom action (webhook) from custom datafield
    - Send custom messages directly from a button (custom datafield)

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

### karoo-tilehunting
- **Extension Name:** [karoo-tilehunting](https://github.com/timklge/karoo-tilehunting)
  - Description: Adds a map overlay that shows explored tiles and an alert if you ride into an unexplored one. Connects to [statshunters.com](https://www.statshunters.com).
  - License: Open Source, Apache 2
  - Features:
    - Shows outlines of explored StatsHunters tiles, biggest square, clusters and unexplored tiles
    - Shows an alert if you ride into an unexplored tile
    - Optionally, shows lines of past activities on the map
    - Data fields that show the number of explored tiles and recently explored tiles

### karoo-routegraph
- **Extension Name:** [karoo-routegraph](https://github.com/timklge/karoo-routegraph)
  - Description: Adds an alternative route elevation profile data field that includes POIs and climbs
  - License: Open Source, Apache 2
  - Features:
    - Colorized route elevation profile with climbs and markers for POIs
    - Vertical / flipped route elevation profile that shows climbs and POIs with their names and remaining distance
    - Data fields that show the remaining elevation and distance to the next POI on the route
   
### Waypoints
- **Extension Name:** [Waypoints](https://github.com/dimskiy/WaypointsKaroo)
  - Description: Waypoints is the extension app improving the navigation convenience: POI online search & offline storage, click-to-navigate.
  - License: Open Source, Apache-2.0
  - Features:
    - Search online for restaurants, shops, cafes etc
    - Open the location of found POIs in the native Karoo navigation app
    - Perform search of spots nearby using device location
    - Support for device-installed SIM cards to perform search with no Wi-Fi available
    - Save POIs for offline use

### karoo-notepad
- **Extension Name:** [karoo-notepad](https://github.com/timklge/karoo-notepad)
  - Description: Minimalistic extension that provides a datafield you can put your own text on
  - License: Open Source, Apache 2
  - Features:
    - Open the extension from the menu to input some custom text
    - Datafield that displays the custom text during riding
    - Additional datafield that can be clicked to open the text editor   

### karoo-spintunes
- **Extension Name:** [karoo-spintunes](https://github.com/timklge/karoo-spintunes)
  - Description: Karoo datafield with media controls for Spotify
  - License: Open Source, Apache 2
  - Features:
    - Play/pause, skip forward, skip backward, shuffle, repeat, add to queue, view and play playlist, view queue
    - Thumbnails for currently played track and playlists
    - Remote control the Spotify app running on your phone (or computer)
    - Offline mode if Spotify is sideloaded on your Karoo (including volume control, automatic volume control based on ride speed)

## Libraries

### ktor-client-karoo

- **Library Name:** [ktor-client-karoo](https://github.com/jonasfranz/ktor-client-karoo)
  - Description: Provides a Ktor Http Client Engine implementation for Karoo
  - License: Open Source, Apache 2
  - Features:
    - Simplifies making http request via the Karoo System Service
    - Works even if the Karoo itself has no wifi connection. (Karoo Companion required)

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
