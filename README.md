# Environment Configuration Package

This package provides generic environment variable handling and configuration loading utilities that can be used across different Go projects.

## Features

- **Environment Variable Parsing**: Parse .env files from bytes or file system
- **Environment Merging**: Merge multiple environment sources with priority
- **Struct Filling**: Use reflection to fill struct fields from environment variables
- **Generic Validation**: Validate configuration structs using struct tags

## Components

### Environment Loading
- `readDotenvBytes()` - Parse .env content from byte arrays
- `getenvs()` - Get all system environment variables
- `mergeEnvMaps()` - Merge multiple environment maps with normalization

### Struct Processing
- `fillStructFromEnv()` - Fill struct fields from environment variables using reflection
- Support for nested structs, slices, and various field types
- Automatic field name detection without requiring tags

### Validation
- Generic struct validation using reflection
- Support for `required`, `min`, `max`, `pattern` struct tags
- Extensible validation with custom field validators

## How Environment Variables Map to Config Fields

Environment variables are automatically mapped to struct fields using the following rules:

1. **Field Name Matching**: Environment variables are normalized to uppercase and matched against struct field names
2. **Nested Structs**: Nested configuration is accessed using underscore notation (e.g., `DATABASE_HOST` maps to `Database.Host`)
3. **Priority Order**: Later environment maps override earlier ones (system env > shared .env > app .env)
4. **Naming Constraints**: Field names must be simple (e.g., `Host`, `Port`) - avoid camelCase like `OneTwo` as it becomes `ONE_TWO` and breaks parsing

## Usage Example

### Configuration Struct
```go
type DatabaseConfig struct {
    Host     string `required:"true"`
    Port     int    `min:"1" max:"65535"`
    Name     string `required:"true"`
    Username string
    Password string
}

type ServerConfig struct {
    Port    int    `required:"true" min:"1" max:"65535"`
    Host    string
    Baseurl string `required:"true"`  // Note: simple field name, not BaseURL
}

type AppConfig struct {
    Environment string         `required:"true"`
    Debug       bool
    Database    DatabaseConfig
    Server      ServerConfig
}
```

### Environment Variables (.env file)
```env
# App-level configuration
ENVIRONMENT=development
DEBUG=true

# Database configuration (nested)
DATABASE_HOST=localhost
DATABASE_PORT=5432
DATABASE_NAME=myapp
DATABASE_USERNAME=user
DATABASE_PASSWORD=secret

# Server configuration
SERVER_PORT=8080
SERVER_HOST=0.0.0.0
SERVER_BASEURL=https://myapp.local
```

### Loading Configuration
```go
package main

import (
    "fmt"
    "reflect"
    
    "github.com/vairogs-go/envconfig"
)

func main() {
    config := &AppConfig{}
    
    // Load environment variables from multiple sources
    envMaps := []map[string]string{
        envconfig.ReadDotenvBytes(sharedEnvContent),  // Shared .env
        envconfig.GetEnvs(),                          // System environment
        envconfig.ReadDotenvBytes(appEnvContent),     // App-specific .env
    }
    
    // Merge all environment sources (later sources override earlier ones)
    merged := envconfig.MergeEnvMaps(envMaps...)
    
    // Fill struct from environment variables (no mapstructure needed)
    err := envconfig.FillStructFromEnv("", reflect.ValueOf(config).Elem(), merged)
    if err != nil {
        panic(fmt.Sprintf("Failed to load config: %v", err))
    }
    
    // Validate the configuration
    validator := envconfig.NewValidator()
    err = validator.ValidateStruct(config)
    if err != nil {
        panic(fmt.Sprintf("Invalid config: %v", err))
    }
    
    // Configuration is now ready to use
    fmt.Printf("Server will run on %s:%d\n", config.Server.Host, config.Server.Port)
    fmt.Printf("Database: %s@%s:%d/%s\n", 
        config.Database.Username, 
        config.Database.Host, 
        config.Database.Port, 
        config.Database.Name)
}
```

### Result
The environment variables automatically populate the configuration struct:
- `ENVIRONMENT=development` → `config.Environment = "development"`
- `DEBUG=true` → `config.Debug = true`
- `DATABASE_HOST=localhost` → `config.Database.Host = "localhost"`
- `DATABASE_PORT=5432` → `config.Database.Port = 5432`
- `SERVER_PORT=8080` → `config.Server.Port = 8080`
- `SERVER_BASEURL=https://myapp.local` → `config.Server.Baseurl = "https://myapp.local"`

This package extracts the generic environment variable handling logic from project-specific configuration, making it reusable across different Go projects.
