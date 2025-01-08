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
    targets: [
        .target(
            name: "NexaSwift",
            dependencies: [
                // .product(name: "llama", package: "llama.cpp")
            ],
            path: "swift/Sources/NexaSwift",
            resources: [
                .copy("lib/libomni_vlm_wrapper_shared.dylib"),
                .copy("lib/libllama.dylib"),
                .copy("lib/libggml_llama.dylib"),
            ],
        ),
        .testTarget(
            name: "NexaSwiftTests",
            dependencies: ["NexaSwift"],
            path: "swift/Tests/NexaSwiftTests"
        )
    ]
)
