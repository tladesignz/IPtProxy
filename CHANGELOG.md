# IPtProxy Changelog

## 5.1.1
- Updated DNSTT to 1.20260125.0.

## 5.1.0
- Added DNSTT Pluggable Transport.
- Slightly improved build script.

## 5.0.0
- Updated Lyrebird to version 0.8.1.
- Improved `OnTransportStopped` to a more versatile `OnTransportEvents`, which reports errors during Snowflake proxy
  search and also its success.

## 4.3.0
- Updated Lyrebird to version 0.8.0.
- Increased minimal iOS version to 15.0.

## 4.2.2
- Expose `SnowflakeProxy`.`EphemeralMinPort` and `EphemeralMaxPort` arguments.

## 4.2.1
- Updated dependencies.
- Added `SnowflakeProxy.NATTypeMeasurementInterval` configuration.
- Support Android 16kB memory alignment.
- Increased minimal Android API to 24.

## 4.2.0
- Updated Lyrebird to version 0.6.1.

## 4.1.2
- Fixed `LyrebirdVersion()` return string.

## 4.1.1
- Failed CocoaPods upload makes version 4.1.0 unusable. Need to use a new version name.

## 4.1.0
- Updated Snowflake to version 2.11.0.
- Updated Lyrebird to version 0.6.0.

## 4.0.1
- Added support of manipulating `PollInterval` of `SnowflakeProxy`.

## 4.0.0
- Complete rewrite of IPtProxy:
  - Got rid of patches and the goptlib interface.
  - Instead, have our own unified code which creates transports using Lyrebird and Snowflake as dependencies.
  - Structured with classes now instead of global functions.
  - Improved interface: 
    - When `#start` returns, it's now safe to use the transport.
    - `#start` will throw errors if something's wrong.
    - Callback for when transport stopped.
- Finally removed Jitpack.
- Updated Snowflake to v2.10.1.
- Updated Lyrebird to v0.5.0.

## 3.8.2
- Reissue because of problems with Maven Central.

## 3.8.1
- Fixed webtunnel support.

## 3.8.0
- Update Lyrebird to v0.2.0. Added Webtunnel support.

## 3.7.0
- Update Snowflake to v2.9.2. SQS arguments changed!

## 3.6.0
- Fixed `StartSnowflake` argument order.

## 3.5.0
- Update Snowflake to v2.9.1. Contains SQS support, hence new arguments.

## 3.4.0
- Update Snowflake to v2.8.1.

## 3.3.0
- Update Snowflake to v2.8.0.

## 3.2.1
- Removed doubled "front" argument, therefore fixed backwards compatibility with IPtProxy 3.

## 3.2.0
- Update Snowflake to v2.7.0.
- Raised minimum needed iOS to 12, since Xcode 15 dropped support for 11.
- Raised minimum needed Android API to 21, since NDK 26.1. dropped support for 19 and 20.

## 3.1.1
- Fixed broken compilation on Apple platforms due to missing library.

## 3.1.0
- Update Snowflake to v2.6.0.

## 3.0.0
- Follow Tor's renaming of the Obfs4proxy fork to Lyrebird. Breaks APIs, hence the 
  huge version jump.

## 2.0.0
- Improved build by stripping paths in output binary, which leak build environment info.
- Log Snowflake Proxy to STDOUT instead of STDERR.
- Fixed event dispatcher crash in recent versions of Snowflake Proxy.
- Switched Obfs4proxy to new fork by Tor Project which has hardened TLS negotiation.
- Fixed security issue with default `StateLocation`. Consumers are now forced to
  define it themselves before first use. Breaking change!

## 1.10.1
- Fixed Snowflake version number.

## 1.10.0
- Updated Snowflake to latest version 2.5.1.

## 1.9.0
- Updated Snowflake to latest version 2.4.1.
- Force netcgo to ensure DNS caching.

## 1.8.1
- Added `Obfs4proxyLogFile` which returns the static log file name of Obfs4proxy.

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