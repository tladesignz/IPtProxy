#
# Be sure to run `pod lib lint IPtProxy.podspec' to ensure this is a
# valid spec before submitting.
#
# Any lines starting with a # are optional, but their use is encouraged
# To learn more about a Podspec see https://guides.cocoapods.org/syntax/podspec.html
#

Pod::Spec.new do |s|
  s.name             = 'IPtProxy'
  s.version          = '1.7.1'
  s.summary          = 'Obfs4proxy and Snowflake Pluggable Transports for iOS and macOS'

  s.description      = <<-DESC
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

    Contained transport versions:

    | Transport  | Version |
    |------------|--------:|
    | Obfs4proxy |  0.0.13 |
    | Snowflake  |   2.3.0 |

                       DESC

  s.homepage         = 'https://github.com/tladesignz/IPtProxy'
  s.license          = { :type => 'MIT', :file => 'LICENSE' }
  s.author           = { 'Benjamin Erhart' => 'berhart@netzarchitekten.com' }
  s.source           = { :git => 'https://github.com/tladesignz/IPtProxy.git', :tag => s.version.to_s }
  s.social_media_url = 'https://twitter.com/tladesignz'

  s.ios.deployment_target = '11.0'
  s.osx.deployment_target = '12'

  s.preserve_paths = 'build.sh', '*.patch', 'IPtProxy.go/*'

  # This will only be executed once.
  s.prepare_command = './build.sh'

  # That's why this is also here, albeit it will be too late here.
  # You will need to re-run `pod update` to make the last line work.
  s.script_phase = {
    :name => 'Go build of IPtProxy.xcframework',
    :execution_position => :before_compile,
    :script => 'sh "$PODS_TARGET_SRCROOT/build.sh"',
  }

  # This will only work, if `prepare_command` was successful, or if you
  # called `pod update` a second time after a build which will have triggered
  # the `script_phase`, or if you ran `build.sh` manually.
  s.vendored_frameworks = "IPtProxy.xcframework"

end
