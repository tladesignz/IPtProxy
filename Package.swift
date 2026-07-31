// swift-tools-version: 6.3
// The swift-tools-version declares the minimum version of Swift required to build this package.

import PackageDescription

let package = Package(
    name: "IPtProxy",
    platforms: [.iOS(.v15), .macOS(.v11)],
    products: [
        .library(
            name: "IPtProxy",
            targets: ["IPtProxy"]
        ),
    ],
    targets: [
        .binaryTarget(
            name: "IPtProxy",
            path: "IPtProxy.xcframework"
        ),
    ],
    swiftLanguageModes: [.v6]
)