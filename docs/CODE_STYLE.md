# Code Style Guide

## Go Code Style

### Package Organization
```
talis/
├── cmd/                    # Command line tools
├── internal/              # Internal packages
│   ├── api/              # API implementation
│   ├── compute/          # Cloud provider implementations
│   ├── db/               # Database layer
│   └── types/            # Common types and interfaces
├── docs/                 # Documentation
└── scripts/              # Utility scripts
```

### Naming Conventions
- Use PascalCase for exported names
- Use camelCase for internal names
- Use snake_case for file names
- Use descriptive package names
- Avoid abbreviations

### Code Organization
- One package per directory
- Package comment at the top of one file
- Logical grouping of related code
- Keep files focused and small

### Error Handling
```go
// Preferred
if err != nil {
    return fmt.Errorf("failed to create instance: %w", err)
}

// Avoid
if err != nil {
    return err
}
```

### Functions and Methods structure
```go
// Preferred
func (s *Service) MethodName(
    arg1 string,
    arg2 int,
) (string, error) {
    // Implementation
}

// Avoid
func (s *Service) MethodName(arg1 string, arg2 int) (string, error) {
    // Implementation
}
```

### Comments and Documentation
- Package comments explain package purpose
- Function comments describe behavior
- Include examples for complex functions
- Document error conditions
- Keep comments up to date

### Testing
- Table-driven tests when possible
- Test error conditions
- Mock external dependencies
- Descriptive test names
- One assertion per test

## Database Conventions

### Table Names
- Use plural, snake_case
- Descriptive but concise
- Include timestamps where appropriate

### Column Names
- Use snake_case
- Be descriptive
- Use proper data types
- Include indexes where needed

### Migrations
- Sequential numbering
- Descriptive names
- Include both up and down
- Test migrations

## Git Conventions

### Commit Messages
```
type(scope): description

[optional body]
[optional footer]
```

Types:
- feat: New feature
- fix: Bug fix
- docs: Documentation
- style: Formatting
- refactor: Code restructuring
- test: Adding tests
- chore: Maintenance

### Branching
- main: Production code
- feature/*: New features
- fix/*: Bug fixes

### Pull Requests
- Clear description
- Reference issues
- Include tests
- Update documentation
- Keep changes focused 