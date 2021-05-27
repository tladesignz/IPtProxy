# IPtProxy Changelog

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