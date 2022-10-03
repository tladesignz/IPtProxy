# IPtProxy Changelog

## 1.8.0
- Updated Obfs4proxy to latest version 0.0.14.
- Updated Snowflake to latest version 2.3.1.
- Added support for macOS 11.
- Fixed warning in Xcode 14.

## 1.7.1
- Fixed Snowflake Proxy support.

## 1.7.0
- Update Snowflake to latest version 2.3.0.
- Added `IPtProxySnowflakeVersion` returning the version of the used Snowflake.

## 1.6.0
- Update Snowflake to latest version 2.2.0.
- Added `IPtProxyObfs4ProxyVersion` returning the version of the used Obfs4proxy.
- Use latest Android NDK v24.0 which raises the minimally supported Android API level to 19.
- Added support for MacOS.

## 1.5.1
- Update Snowflake to latest main. Contains a crash fix.
- Added `IsSnowflakeProxyRunning` method to easily check,
  if the Snowflake Proxy is running.
- Exposed `IsPortAvailable` so consumers don't need to 
  implement this themselves, if they happen to do something similar.

## 1.5.0
- Updated Obfs4proxy to latest version 0.0.13.
- Updated Snowflake to latest version 2.1.0.
- Fixed bug when stopping Snowflake proxy. (Thanks bitmold!)

## 1.4.0
- Updated Obfs4proxy to latest 0.0.13-dev which fixes a bug which made prior 
  versions distinguishable.
- Fixed minor documentation issues.

## 1.3.0
- Updated Snowflake to version 2.0.1.
- Added Snowflake AMP support.
- Switched to newer DTLS library, which improves fingerprinting resistance for Snowflake.
- Added callback to `StartSnowflakeProxy`, to allow counting of connected clients.
- Fixed iOS warnings about wrong iOS SDK.

## 1.2.0
- Added explicit support for a proxy behind Obfs4proxy.

## 1.1.0
- Updated Snowflake to latest master. Fixes multiple minor issues.
- Registers Snowflake proxy with type "iptproxy" to improve statistics.

## 1.0.0
- Updated Snowflake to latest master. Fixes multiple minor issues.
- Updated Obfs4proxy to latest master. Contains a minor fix for Meek.
- Added port test mechanism to avoid port collisions when started multiple times.
- Improved documentation.

## 0.6.0
- Updated Obfs4proxy to latest master. Fixes support for unsafe logging.
- Added `StopSnowflakeProxy` from feature branch.
- Updated Snowflake to latest master. Fixes multiple minor issues.

## 0.5.2
- Updated Obfs4proxy to fix broken meek_lite due to Microsoft Azure certificate
  changes. NOTE: If you still experience HPKP issues, you can use 
  "disableHPKP=true" in the meek_lite configuration.

## 0.5.1

- Base on latest Snowflake master which contains a lot of patches we previously
  had to provide ourselves.

## 0.5.0

- Added `StopSnowflake` function.

## 0.4.0

- Added `StopObfs4Proxy` function.
- Updated Snowflake to latest master.

## 0.3.0

- Added Snowflake Proxy support, so contributors can run proxies on their 
  mobile devices.
- Updated Snowflake to latest master.
- Fixed doc to resemble proper Objective-C documentation.

## 0.2.0

- Improved Android support.
- Improved documentation.
- Updated Snowflake to latest master.

## 0.1.0

Initial version