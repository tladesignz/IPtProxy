# IPtProxy

Obfs4proxy and Snowflake Pluggable Transports for iOS, MacOS and Android

[![JitPack](https://jitpack.io/v/tladesignz/IPtProxy.svg)](https://jitpack.io/#tladesignz/IPtProxy)
[![Version](https://img.shields.io/cocoapods/v/IPtProxy.svg?style=flat)](https://cocoapods.org/pods/IPtProxy)
[![License](https://img.shields.io/cocoapods/l/IPtProxy.svg?style=flat)](https://cocoapods.org/pods/IPtProxy)
[![Platform](https://img.shields.io/cocoapods/p/IPtProxy.svg?style=flat)](https://cocoapods.org/pods/IPtProxy)

| Transport  | Version |
|------------|--------:|
| Obfs4proxy |  0.0.13 |
| Snowflake  |   2.3.0 |

Both Obfs4proxy and Snowflake Pluggable Transports are written in Go, which
is a little annoying to use on iOS and Android.
This project encapsulates all the machinations to make it work and provides an
easy to install binary including a wrapper around both.

Problems solved in particular are:

- One cannot compile `main` packages with `gomobile`. Both PTs are patched
  to avoid this.
- Both PTs are gathered under one roof here, since you cannot have two
  `gomobile` frameworks as dependencies, since there are some common Go
  runtime functions exported, which would create a name clash.
- Environment variable changes during runtime will not be recognized by
  `goptlib` when done from within Swift/Objective-C. Therefore, sensible
  values are hardcoded in the Go wrapper.
- Snowflake and Obfs4proxy are patched to accept all configuration parameters
  directly.
- Free ports to be used are automatically found by this library and returned to the
  consuming app. You can use the initial values for premature configuration just
  fine in situations, where you can be pretty sure, they're going to be available
  (typically on iOS). When that's not the case (e.g. multiple instances of your app
  on a multi-user Android), you should first start the transports and then use the 
  returned ports for configuration of other components (e.g. Tor). 

## iOS/macOS

### Installation

IPtProxy is available through [CocoaPods](https://cocoapods.org). To install
it, simply add the following line to your `Podfile`:

```ruby
pod 'IPtProxy', '~> 1.7'
```

### Getting Started

There's a companion library [IPtProxyUI](https://github.com/tladesignz/IPtProxyUI)
which explains the use of IPtProxy and provides all the necessary UI and additional 
information to use this library completely in a Tor context.


## Android 

### Installation

IPtProxy is available through [JitPack](https://jitpack.io). To install
it, simply add the following line to your `build.gradle` file:

```groovy
implementation 'com.github.tladesignz:IPtProxy:1.7.1'
```

And this to your root `build.gradle` at the end of repositories:

```groovy
allprojects {
	repositories {
		// ...
		maven { url 'https://jitpack.io' }
	}
}
```

For newer Android Studio projects created in 
[Android Studio Bumblebee | 2021.1.1](https://developer.android.com/studio/preview/features?hl=hu#settings-gradle) 
or newer</a>, the JitPack repository needs to be added into the root level file `settings.gradle` 
instead of `build.gradle`:

```groovy
dependencyResolutionManagement {
    repositoriesMode.set(RepositoriesMode.FAIL_ON_PROJECT_REPOS)
    repositories {
	  // ...
        maven { url 'https://jitpack.io' }
    }
}
```

### Getting Started

If you are building a new Android application be sure to declare that it uses the
`INTERNET` permission in your Android Manifest:

```xml
<?xml version="1.0" encoding="utf-8"?>
<manifest xmlns:android="http://schemas.android.com/apk/res/android"
    package="my.test.app">

    <uses-permission android:name="android.permission.INTERNET"/>
    <application ...

```

Before using IPtProxy you need to specify a place on disk for it to store its state
information. We recommend the path returned by `Context#getCacheDir()`:

```java
File fileCacheDir = new File(getCacheDir(), "pt");

if (!fileCacheDir.exists()) fileCacheDir.mkdir();

IPtProxy.setStateLocation(fileCacheDir.getAbsolutePath());
```


## Build

### Requirements

This repository contains a precompiled iOS and Android version of IPtProxy.
If you want to compile it yourself, you'll need Go 1.16 as a prerequisite.

You will also need Xcode installed when compiling for iOS and an Android NDK
when compiling for Android.

The build script needs the gomobile binary and will install it, if not available, yet.
However, you'll still need to make it accessible in your `$PATH`.

So, if it's not already, add `$GOPATH/bin` to `$PATH`. The default location 
for `$GOPATH` is `$HOME/go`: 

```bash
export PATH=$HOME/go/bin/:$PATH` 
```

### iOS

Make sure Xcode and Xcode's command line tools are installed. Then run

```bash
rm -rf IPtProxy.xcframework && ./build.sh
```

This will create an `IPtProxy.xcframework`, which you can directly drop in your app,
if you don't want to rely on CocoaPods.

### Android

Make sure that `javac` is in your `$PATH`. If you do not have a JDK instance, on Debian systems you can install it with: 

```bash
apt install default-jdk 
````

If they aren't already, make sure the `$ANDROID_HOME` and `$ANDROID_NDK_HOME` 
environment variables are set:

```bash
export ANDROID_HOME=~/Android/Sdk
export ANDROID_NDK_HOME=$ANDROID_HOME/ndk/$NDK_VERSION

rm -rf IPtProxy.aar IPtProxy-sources.jar && ./build.sh android
```

This will create an `IPtProxy.aar` file, which you can directly drop in your app, 
if you don't want to rely on JitPack.

On certain CPU architectures `gobind` might fail with this error due to setting
a flag that is no longer supported by Go 1.16:

```
go tool compile: exit status 1
unsupported setting GO386=387. Consider using GO386=softfloat instead.
gomobile: go build -v -buildmode=c-shared -o=/tmp/gomobile-work-855414073/android/src/main/jniLibs/x86/libgojni.so ./gobind failed: exit status 1
```

If this is the case, you will need to set this flag to build IPtProxy:

```bash
export GO386=sse2
``` 


## Authors

- Benjamin Erhart, berhart@netzarchitekten.com
- Nathan Freitas
- Bim

for the Guardian Project https://guardianproject.info

## License

IPtProxy is available under the MIT license. See the [LICENSE](LICENSE) file for more info.
