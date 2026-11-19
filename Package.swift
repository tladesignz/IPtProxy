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
            url: "https://github.com/tladesignz/IPtProxy/releases/download/5.9.0/IPtProxy.xcframework.zip",
            checksum: "8aea53dfd7eec8045dfc0e5f58ca328d444233258f9892f3ad381733a0763dfa"
        ),
    ],
    swiftLanguageModes: [.v6]
)