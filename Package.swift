// swift-tools-version: 6.3
// The swift-tools-version declares the minimum version of Swift required to build this package.

import PackageDescription

let package = Package(
    name: "IPtProxy",
    platforms: [.iOS(.v15), .macOS(.v12)],
    products: [
        .library(
            name: "IPtProxy",
            targets: ["IPtProxy"]
        ),
    ],
    targets: [
        .binaryTarget(
            name: "IPtProxy",
            url: "https://github.com/tladesignz/IPtProxy/releases/download/5.8.1/IPtProxy.xcframework.zip",
            checksum: "8218a706709fded3c783f096b583d48c52c289712595879b0a07c684eaca9220"
        ),
    ],
    swiftLanguageModes: [.v6]
)