// swift-tools-version: 6.0

import PackageDescription

let package = Package(
    name: "NexaSwift",
    platforms: [
        .macOS(.v15),
        .iOS(.v18),
        .watchOS(.v11),
        .tvOS(.v18),
        .visionOS(.v2)
    ],
    products: [
        .library(name: "NexaSwift", targets: ["NexaSwift"]),
    ],
    dependencies: [
        .package(url: "https://github.com/ggerganov/llama.cpp.git", branch: "master")
    ],
    targets: [
        .target(
            name: "NexaSwift", 
            dependencies: [
                .product(name: "llama", package: "llama.cpp")
            ],
            path: "swift/Sources/NexaSwift"),
        .testTarget(
            name: "NexaSwiftTests", 
            dependencies: ["NexaSwift"],
            path: "swift/Tests/NexaSwiftTests"),
    ]
)
