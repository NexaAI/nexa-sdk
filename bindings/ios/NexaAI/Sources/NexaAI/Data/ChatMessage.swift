// Copyright (c) 2025 Nexa AI
//
// LICENSE NOTICE - DUAL LICENSING:
// - NPU models and inference: CC-BY-NC 4.0 (NON-COMMERCIAL USE ONLY)
// - GPU/CPU models and inference: Apache 2.0 (FREE FOR ALL USE)

import Foundation

public struct ChatMessage {
    public var role: Role
    public var content: String
    public var images: [String]
    public var audios: [String]

    public init(
        role: Role,
        content: String,
        images: [String] = [],
        audios: [String] = []
    ) {
        self.role = role
        self.content = content
        self.images = images
        self.audios = audios
    }
}

public enum Role: String {
    case user
    case assistant
    case system
}
