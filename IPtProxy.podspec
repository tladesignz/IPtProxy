#
# Be sure to run `pod lib lint IPtProxy.podspec' to ensure this is a
# valid spec before submitting.
#
# Any lines starting with a # are optional, but their use is encouraged
# To learn more about a Podspec see https://guides.cocoapods.org/syntax/podspec.html
#

Pod::Spec.new do |s|
  s.name             = 'IPtProxy'
  s.version          = '5.1.0'
  s.summary          = 'Lyrebird/Obfs4proxy, Snowflake and DNSTT Pluggable Transports for iOS and macOS'

  s.description      = <<-DESC
    Lyrebird/Obfs4proxy as well as Snowflake and DNSTT Pluggable Transports are written in Go, which
    is a little annoying to use on iOS and Android.
    This project encapsulates all the machinations to make it work and provides an
    easy-to-install binary including a wrapper around both.

    Problems solved in particular are:

    - One cannot compile `main` packages with `gomobile`. For all PTs, IPtProxy
      provides wrapper code to use them as libraries.
    - All PTs are gathered under one roof here, since you cannot have two
      `gomobile` frameworks as dependencies. There are some common Go
      runtime functions exported, which would create a name clash.
    - Free ports to be used are automatically found by this library and can be fetched
      by the consuming app after start.

    Contained transport versions:

    | Transport | Version      |
    |-----------|--------------|
    | Lyrebird  | 0.8.1        |
    | Snowflake | 2.11.0       |
    | DNSTT     | 1.20240513.0 |

                       DESC

  s.homepage         = 'https://github.com/tladesignz/IPtProxy'
  s.license          = { :type => 'MIT', :file => 'LICENSE' }
  s.author           = { 'Benjamin Erhart' => 'berhart@netzarchitekten.com' }
  s.source           = { :git => 'https://github.com/tladesignz/IPtProxy.git', :tag => s.version.to_s }
  s.social_media_url = 'https://chaos.social/@tla'

  s.ios.deployment_target = '15.0'
  s.osx.deployment_target = '11.0'

  s.preserve_paths = 'build.sh', '*.patch', 'IPtProxy.go/*'

  # This will only be executed once.
  s.prepare_command = './build.sh'

  # That's why this is also here, albeit it will be too late here.
  # You will need to re-run `pod update` to make the last line work.
  s.script_phase = {
    :name => 'Go build of IPtProxy.xcframework',
    :execution_position => :before_compile,
    :script => 'sh "$PODS_TARGET_SRCROOT/build.sh"',
    :output_files => ['$(DERIVED_FILE_DIR)/IPtProxy.xcframework'],
  }

  # This will only work, if `prepare_command` was successful, or if you
  # called `pod update` a second time after a build which will have triggered
  # the `script_phase`, or if you ran `build.sh` manually.
  s.vendored_frameworks = 'IPtProxy.xcframework'

  s.libraries = 'resolv'

end
