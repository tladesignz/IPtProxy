# IPtProxy

Obfs4proxy and Snowflake Pluggable Transports for iOS (and Android!)

[![Version](https://img.shields.io/cocoapods/v/IPtProxy.svg?style=flat)](https://cocoapods.org/pods/IPtProxy)
[![License](https://img.shields.io/cocoapods/l/IPtProxy.svg?style=flat)](https://cocoapods.org/pods/IPtProxy)
[![Platform](https://img.shields.io/cocoapods/p/IPtProxy.svg?style=flat)](https://cocoapods.org/pods/IPtProxy)

Both Obfs4proxy and Snowflake Pluggable Transports are written in Go, which
is a little annoying to use on iOS.
This pod encapsulates all the machinations to make it work and provides an
easy to install binary including a wrapper around both.

Problems solved in particular are:

- One cannot compile `main` packages with `gomobile`. Both PTs are patched
  to avoid this.
- Both PTs are gathered under one roof here, since you cannot have two
  `gomobile` frameworks in your iOS code, since there are some common Go
  runtime functions exported, which will create a name clash.
- Environment variable changes during runtime will not be recognized by
  `goptlib` when done from within Swift/Objective-C. Therefore, sensible
  values are hardcoded in the Go wrapper.
- The ports where the PTs will listen on are hardcoded, since communicating
  the used ports back to the app would be quite some work (e.g. trying to
  read it from STDOUT) for very little benefit.
- Snowflake currently can only be configured via command line, not via the
  PT spec's method of using SOCKS username and password arguments.
  Therefore Snowflake is patched to accept arguments via its `Main` method.

Both PTs are contained at their latest `master` commit, as per 2020-09-10.


## Requirements

This repository contains a precompiled iOS version of IPtProxy.
If you want to compile it yourself, you'll need Go 1.15 as a prerequisite.

You will also need Xcode installed when compiling for iOS and an Android NDK
when compiling for Android.

## Installation (iOS)

IPtProxy is available through [CocoaPods](https://cocoapods.org). To install
it, simply add the following line to your Podfile:

```ruby
pod 'IPtProxy', '~> 0.1'
```

## Installation (Android)

Use `build.sh android` to compile an AAR for Android. You can directly drop that
in your app.

## Author

Benjamin Erhart, berhart@netzarchitekten.com
for the Guardian Project https://guardianproject.info

## License

IPtProxy is available under the MIT license. See the LICENSE file for more info.
