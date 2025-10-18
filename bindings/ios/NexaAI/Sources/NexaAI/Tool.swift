
/**
/** Tool definition */
typedef struct {
    const char* name;            /** name of the function */
    const char* description;     /** description of the function in natural language */
    const char* parameters_json; /** JSON schema for the function parameters */
} ml_ToolFunction;

typedef struct {
    const char*            type;     /** always "function" */
    const ml_ToolFunction* function; /** pointer to ToolFunction */
} ml_Tool;
 */

// Copyright (c) 2025 Nexa AI
//
// LICENSE NOTICE - DUAL LICENSING:
// - NPU models and inference: CC-BY-NC 4.0 (NON-COMMERCIAL USE ONLY)
// - GPU/CPU models and inference: Apache 2.0 (FREE FOR ALL USE)

import Foundation

public struct Tool {

    public enum ToolType: String {
        case function = "function"
    }

    public struct Function {
        /// name of the function
        public let name: String
        /// description of the function in natural language
        public let description: String
        /// JSON schema for the function parameters
        public let parameters: String
        public init(name: String, description: String, parameters: String) {
            self.name = name
            self.description = description
            self.parameters = parameters
        }
    }

    /// always "function"
    public let type: ToolType

    public let function: Function

    public init(type: ToolType = .function, function: Function) {
        self.type = type
        self.function = function
    }
}
