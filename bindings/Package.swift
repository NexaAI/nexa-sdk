// swift-tools-version: 5.9
// The swift-tools-version declares the minimum version of Swift required to build this package.

import PackageDescription

let package = Package(
    name: "NexaAI",
    platforms: [
        .macOS(.v14),.iOS(.v17)
    ],
    products: [
        // Products define the executables and libraries a package produces, making them visible to other packages.
        .library(
            name: "NexaAI",
            targets: ["NexaAI"]),
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
                .product(name: "nexa_bridge", package: "nexasdk-mobile-iOS-framework"),
                .product(name: "llama", package: "nexasdk-mobile-iOS-framework"),
                .product(name: "common", package: "nexasdk-mobile-iOS-framework"),
                .product(name: "mtmd", package: "nexasdk-mobile-iOS-framework")
            ],
            path: "Sources",
            swiftSettings: [.interoperabilityMode(.Cxx)]
        ),
        .testTarget(
            name: "NexaAITests",
            dependencies: ["NexaAI"],
            swiftSettings: [.interoperabilityMode(.Cxx)]
            // if you want test on iOS/sim, uncomment the line below to add resources
//            resources: [
//                .copy("modelfiles/jina-embeddings-v2-base-en-Q4_0.gguf"),
//                .copy("modelfiles/Qwen3-4B-Q4_K_M.F32.gguf"),
//                .copy("modelfiles/jina-reranker-v2-base-multilingual.F16.gguf"),
//                .copy("modelfiles/jina_rerank_tokenizer.json"),
//                .copy("modelfiles/test-2.mp3"),
//                .copy("modelfiles/test1.jpg"),
//                .copy("modelfiles/bge-reranker-v2-m3-Q4_0.gguf")
//            ]
        ),
    ]
)
