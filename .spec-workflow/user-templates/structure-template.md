# Project Structure

## ディレクトリ構成

```
[Define your project's directory structure. Examples below - adapt to your project type]

Example for a library/package:
project-root/
├── src/                    # Source code
├── tests/                  # Test files
├── docs/                   # Documentation
├── examples/               # Usage examples
└── [build/dist/out]        # Build output

Example for an application:
project-root/
├── [src/app/lib]           # Main source code
├── [assets/resources]      # Static resources
├── [config/settings]       # Configuration
├── [scripts/tools]         # Build/utility scripts
└── [tests/spec]            # Test files

Common patterns:
- Group by feature/module
- Group by layer (UI, business logic, data)
- Group by type (models, controllers, views)
- Flat structure for simple projects
```

## 命名規則

### ファイル名・ディレクトリ名

- **Components/Modules**: [e.g., `PascalCase`, `snake_case`, `kebab-case`]
- **Services/Handlers**: [e.g., `UserService`, `user_service`, `user-service`]
- **Utilities/Helpers**: [e.g., `dateUtils`, `date_utils`, `date-utils`]
- **Tests**: [e.g., `[filename]_test`, `[filename].test`, `[filename]Test`]

### コード内

- **Classes/Types**: [e.g., `PascalCase`, `CamelCase`, `snake_case`]
- **Functions/Methods**: [e.g., `camelCase`, `snake_case`, `PascalCase`]
- **Constants**: [e.g., `UPPER_SNAKE_CASE`, `SCREAMING_CASE`, `PascalCase`]
- **Variables**: [e.g., `camelCase`, `snake_case`, `lowercase`]

## Code Organization Principles

1. **Single Responsibility**: Each file should have one clear purpose
2. **Modularity**: Code should be organized into reusable modules
3. **Testability**: Structure code to be easily testable
4. **Consistency**: Follow patterns established in the codebase

## Module Boundaries

[Define how different parts of your project interact and maintain separation of concerns]

Examples of boundary patterns:

- **Core vs Plugins**: Core functionality vs extensible plugins
- **Public API vs Internal**: What's exposed vs implementation details
- **Platform-specific vs Cross-platform**: OS-specific code isolation
- **Stable vs Experimental**: Production code vs experimental features
- **Dependencies direction**: Which modules can depend on which

## Code Size Guidelines

[Define your project's guidelines for file and function sizes]

Suggested guidelines:

- **File size**: [Define maximum lines per file]
- **Function/Method size**: [Define maximum lines per function]
- **Class/Module complexity**: [Define complexity limits]
- **Nesting depth**: [Maximum nesting levels]

## Documentation Standards

- All public APIs must have documentation
- Complex logic should include inline comments
- README files for major modules
- Follow language-specific documentation conventions
