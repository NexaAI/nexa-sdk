// swift-tools-version: 5.9
// The swift-tools-version declares the minimum version of Swift required to build this package.
import PackageDescription

let package = Package(
    name: "NexaAI",
    platforms: [
        .macOS(.v14), .iOS(.v17),
    ],
    products: [
        .library(name: "NexaAI", targets: ["NexaAI"])
    ],
    dependencies: [
        .package(url: "git@github.com:NexaAI/nexasdk-mobile-iOS-framework.git", branch: "main")
    ],
    targets: [
        // Targets are the basic building blocks of a package, defining a module or a test suite.
        // Targets can depend on other targets in this package and products from dependencies.
        .target(
            name: "NexaAI",
            dependencies: [
                .product(name: "NexaBridge", package: "nexasdk-mobile-iOS-framework"),
                .product(name: "LlamaPlugin", package: "nexasdk-mobile-iOS-framework"),
                .product(name: "llama", package: "nexasdk-mobile-iOS-framework"),
                .product(name: "common", package: "nexasdk-mobile-iOS-framework"),
                .product(name: "mtmd", package: "nexasdk-mobile-iOS-framework"),
                .product(name: "ggml", package: "nexasdk-mobile-iOS-framework"),
                .product(name: "ggml-base", package: "nexasdk-mobile-iOS-framework"),
                .product(name: "ggml-cpu", package: "nexasdk-mobile-iOS-framework"),
                .product(name: "ggml-metal", package: "nexasdk-mobile-iOS-framework")
            ],
            path: "bindings/ios/NexaAI/Sources",
            swiftSettings: [.interoperabilityMode(.Cxx)]
        ),
        .testTarget(
            name: "NexaAITests",
            dependencies: ["NexaAI"],
            path: "bindings/ios/NexaAI/Tests",
            swiftSettings: [.interoperabilityMode(.Cxx)],
        ),
    ]
)
