# IPtProxy

Obfs4proxy and Snowflake Pluggable Transports for iOS and Android

[![JitPack](https://jitpack.io/v/tladesignz/IPtProxy.svg)](https://jitpack.io/#tladesignz/IPtProxy)
[![Version](https://img.shields.io/cocoapods/v/IPtProxy.svg?style=flat)](https://cocoapods.org/pods/IPtProxy)
[![License](https://img.shields.io/cocoapods/l/IPtProxy.svg?style=flat)](https://cocoapods.org/pods/IPtProxy)
[![Platform](https://img.shields.io/cocoapods/p/IPtProxy.svg?style=flat)](https://cocoapods.org/pods/IPtProxy)

Both Obfs4proxy and Snowflake Pluggable Transports are written in Go, which
is a little annoying to use on iOS and Android.
This project encapsulates all the machinations to make it work and provides an
easy to install binary including a wrapper around both.

Problems solved in particular are:

- One cannot compile `main` packages with `gomobile`. Both PTs are patched
  to avoid this.
- Both PTs are gathered under one roof here, since you cannot have two
  `gomobile` frameworks as dependencies, since there are some common Go
  runtime functions exported, which will create a name clash.
- Environment variable changes during runtime will not be recognized by
  `goptlib` when done from within Swift/Objective-C. Therefore, sensible
  values are hardcoded in the Go wrapper.
- The ports where the PTs will listen on are hardcoded, since communicating
  the used ports back to the app would be quite some work (e.g. trying to
  read it from STDOUT) for very little benefit.
- Snowflake and Obfs4proxy are patched to accept all configuration parameters 
  directly.

Both PTs are contained at their latest `master` commit, as per 2020-11-26.

## iOS Installation

IPtProxy is available through [CocoaPods](https://cocoapods.org). To install
it, simply add the following line to your `Podfile`:

```ruby
pod 'IPtProxy', '~> 0.5'
```

## Android Installation

IPtProxy is available through [JitPack](https://jitpack.io). To install
it, simply add the following line to your `build.gradle` file:

```groovy
implementation 'com.github.tladesignz:IPtProxy:0.5.1'
```

And this to your root `build.gradle` at the end of repositories:

```groovy
allprojects {
	repositories {
		...
		maven { url 'https://jitpack.io' }
	}
}
```

## Build

### Requirements

This repository contains a precompiled iOS and Android version of IPtProxy.
If you want to compile it yourself, you'll need Go 1.15 as a prerequisite.

You will also need Xcode installed when compiling for iOS and an Android NDK
when compiling for Android.

If it's not already, add `$GOPATH/bin` to `$PATH`. The default location 
for `$GOPATH` is `$HOME/go` 

```bash
export PATH=$HOME/go/bin/:$PATH` 
```

### iOS

Make sure Xcode and Xcode's command line tools are installed. Then run

```bash
rm -rf IPtProxy.framework && ./build.sh
```

This will create an `IPtProxy.framework`, which you can directly drop in your app,
if you don't want to rely on CocoaPods.

### Android

If they aren't already, make sure the `$ANDROID_HOME` and `$ANDROID_NDK_HOME` 
environment variables are set:

```bash
export ANDROID_HOME=~/Android/Sdk
export ANDROID_NDK_HOME=$ANDROID_HOME/ndk/$NDK_VERSION

rm -rf IPtProxy.aar IPtProxy-sources.jar && ./build.sh android
```

This will create an `IPtProxy.aar` file, which you can directly drop in your app, 
if you don't want to rely on JitPack.

## Authors

- Benjamin Erhart, berhart@netzarchitekten.com
- Nathan Freitas
- bitmold

for the Guardian Project https://guardianproject.info

## License

IPtProxy is available under the MIT license. See the LICENSE file for more info.
