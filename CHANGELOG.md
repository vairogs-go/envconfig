# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- Initial release of envconfig package
- Environment variable parsing from .env files with `ReadDotenvBytes()`
- System environment variable retrieval with `GetEnvs()`
- Environment map merging with priority support via `MergeEnvMaps()`
- Struct field population from environment variables using `FillStructFromEnv()`
- Configuration validation with struct tags via `StructValidator`
- Support for nested structs with dot notation (e.g., `DATABASE_HOST` → `Database.Host`)
- String slice support with comma-separated values
- Validation tags: `required`, `min`, `max`, `pattern`
- Pattern validation for alphanumeric strings
- Comprehensive error handling with `ValidationError` and `ValidationErrors` types
- Anonymous/embedded struct support
- Pointer dereferencing for nested structs
- Complete package documentation with usage examples
- Extensive test coverage (23 test functions, 83 sub-tests)

### Technical Details
- Module name: `github.com/vairogs-go/envconfig`
- Go version: 1.24+
- Package name: `envconfig`
- Zero external dependencies
- Reflection-based struct field mapping
- Automatic field name detection (no tags required)
- Environment variable normalization (uppercase, underscore to dot conversion)
- Thread-safe operations

### Documentation
- Complete API documentation with examples
- README with usage guide and environment variable mapping rules
- Comprehensive test suite demonstrating all functionality