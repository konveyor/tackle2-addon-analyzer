# Resource Injector

The Resource Injector is a system that dynamically injects Hub resources (identities and settings) into provider configurations at runtime. This allows provider extensions to reference credentials, settings, and other resources from the Tackle2 Hub without hardcoding values.

## Overview

The injector system consists of two main components:

1. **Injector** (`cmd/injector.go`): Base variable replacement engine
2. **ResourceInjector** (`cmd/injector.go`): Fetches resources from Hub and builds the injection dictionary

## How It Works

### 1. Extension Metadata

Extensions define resources they need in their metadata using a declarative format:

```yaml
metadata:
  provider:
    address: localhost:$(PORT)
    initConfig:
    - providerSpecificConfig:
        mavenInsecure: $(maven.insecure)
        mavenSettingsFile: $(maven.settings.path)
      name: java
  resources:
  - selector: identity:kind=maven
    fields:
    - key: maven.settings.path
      name: settings
      path: /shared/creds/maven/settings.xml
  - selector: setting:key=mvn.insecure.enabled
    fields:
    - key: maven.insecure
      name: value
```

### 2. Resource Selection

Resources are identified using selectors with the format: `[namespace/]kind:name=value`

**Supported Resource Kinds:**

- `identity` - Selects credentials/identities from Hub
- `setting` - Selects settings from Hub

**Examples:**
- `identity:kind=maven` - Find identity where kind=maven
- `setting:key=mvn.insecure.enabled` - Get setting with key=mvn.insecure.enabled

### 3. Field Extraction

For each resource, fields define what data to extract and how to inject it:

```go
type Field struct {
    Name    string // Field name in the resource object
    Path    string // Optional: write value to file at this path
    Key     string // Dictionary key for variable substitution
    Type    string // Optional: cast type (string, integer, boolean)
    Default any    // Optional: default value if resource not found
}
```

**Field Processing:**

- **Without `path`**: Value is cast to the specified type and stored in the dictionary as-is
- **With `path`**: Value is written to a file, and the file path is stored in the dictionary

### 4. Variable Substitution

Variables use the format `$(variable)` and are replaced with dictionary values:

```yaml
mavenSettingsFile: $(maven.settings.path)
```

The injector recursively processes:
- Strings: Replaces `$(key)` with dictionary values
- Maps: Processes all values
- Arrays: Processes all elements

**Substitution Rules:**

- If the entire string is a variable (e.g., `"$(key)"`), the original type is preserved
- If the variable is part of a larger string (e.g., `"prefix-$(key)-suffix"`), result is always a string
- Multiple variables in one string are all replaced

### 5. Integration Flow

The `Settings.AppendExtensions()` method in `cmd/settings.go` orchestrates the injection:

```go
func (r *Settings) AppendExtensions(mode *Mode) (err error) {
    // 1. Load extensions from Hub
    addon, err := addon.Addon(true)

    for _, extension := range addon.Extensions {
        // 2. Extract metadata
        md, err = r.metadata(&extension)

        // 3. Skip if provider already exists
        if r.hasProvider(&md.Provider) {
            continue
        }

        // 4. Inject builtin values (e.g., location)
        builtin := r.injectBuiltins(md, mode)

        // 5. Create ResourceInjector with builtins
        injector := ResourceInjector{}
        injector.Use(builtin)

        // 6. Fetch resources from Hub and inject
        err = injector.Inject(md)

        // 7. Add processed provider to settings
        r.content = append(r.content, md.Provider)
    }
    return
}
```

## Implementation Details

### ResourceInjector.Inject()

The injection process happens in phases:

**Phase 1: Build Dictionary**
```go
func (r *ResourceInjector) build(md *Metadata) (err error) {
    // Add defaults first
    for _, resource := range md.Resources {
        err = r.addDefaults(&resource)
    }

    // Fetch and add resources
    for _, resource := range md.Resources {
        parsed := ParsedSelector{}
        parsed.With(resource.Selector)

        switch strings.ToLower(parsed.kind) {
        case "identity":
            // Search for identity by kind (direct/indirect)
            identity, found, err := addon.Application.Select(...).Identity.
                Decrypted().Search().
                Direct(parsed.value).
                Indirect(parsed.value).
                Find()
            if found {
                err = r.add(&resource, identity)
            }

        case "setting":
            // Get setting by key
            setting := &api.Setting{}
            err = addon.Setting.Get(parsed.value, &setting.Value)
            err = r.add(&resource, setting)
        }
    }
}
```

**Phase 2: Variable Replacement**
```go
func (r *Injector) inject(in any) (out any) {
    // Recursively walks data structure
    // Replaces all $(variable) references with dictionary values
}
```

### Field Value Processing

When adding a field to the dictionary:

1. **File Writing** (if `path` is specified):
   ```go
   if f.Path != "" {
       err = r.write(f.Path, v)  // Write value to file
       v = f.Path                // Store path in dictionary
   }
   ```

2. **Type Casting** (if `type` is specified):
   - `string`: Convert using `fmt.Sprintf("%v", object)`
   - `integer`: Convert from int/bool/string
   - `boolean`: Convert from bool/int/string

3. **Conflict Detection**:
   - Keys cannot be redefined
   - Raises `KeyConflictError` if attempted

## Error Handling

The injector defines custom errors for specific failure cases:

- `SelectorNotSupported`: Unknown resource kind
- `FieldNotMatched`: Field not found in resource object
- `TypeError`: Type casting failed
- `KeyConflictError`: Dictionary key redefined

## Builtin Variables

The system injects builtin variables automatically:

- `builtin.location`: The analysis location (source code path)

## Example Workflow

Given this extension metadata:

```yaml
resources:
- selector: identity:kind=maven
  fields:
  - key: maven.settings.path
    name: settings
    path: /shared/creds/maven/settings.xml
```

The system:

1. Fetches identity from Hub where `kind=maven`
2. Extracts the `settings` field value from the identity
3. Writes the settings content to `/shared/creds/maven/settings.xml`
4. Stores `maven.settings.path = "/shared/creds/maven/settings.xml"` in dictionary
5. Replaces all `$(maven.settings.path)` references with the file path

## Key Design Principles

1. **Declarative**: Resources are specified in extension metadata, not code
2. **Type-Safe**: Supports type casting with validation
3. **File Management**: Automatically writes credential files when needed
4. **Conflict Detection**: Prevents accidental key overwrites
5. **Recursive**: Handles nested structures (maps, arrays)
6. **Preserve Types**: When entire value is a variable, original type is maintained
