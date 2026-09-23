# Testing Guide for karpenter-provider-vsphere

## Test Constants & Fixtures Strategy

### Core Principle
**Don't create a giant `constants.go` for all tests** — it becomes a maintenance nightmare.

### Guidelines

#### 1. Keep constants close to where they're used
- Constants used by only one package → **package-local const block**
- Example: `test/e2e/finder_test.go` has finder-only test constants
- **Benefit**: Keeps tests readable, avoids cross-repo navigation

#### 2. Only put truly global test defaults in `internal/testutil`
- For values reused across many packages
- **Prefer default objects over default constants**
- Example:
```go
// Better than constants
func DefaultNodeClass() *v1alpha1.VsphereNodeClass {
    return &v1alpha1.VsphereNodeClass{
        Name: "default",
        // ... sensible defaults
    }
}
```

#### 3. Prefer builders over constants
Instead of this:
```go
const (
    TestCPU    = 8
    TestMemory = "32Gi"
    TestPrice  = 0.42
)
```

Use builders:
```go
func NewInstanceType() *cloudprovider.InstanceType {
    return &cloudprovider.InstanceType{
        Name: "default",
        Capacity: v1.ResourceList{
            corev1.ResourceCPU:    resource.MustParse("8"),
            corev1.ResourceMemory: resource.MustParse("32Gi"),
        },
    }
}
```

Then in tests:
```go
it := testutil.NewInstanceType().
    WithCPU("16").
    WithMemory("64Gi").
    Build()
```

**Benefits**: 
- If defaults change, one place updates everything
- More readable than named constants
- Enables fluent chaining for test customization

#### 4. Never centralize values that reduce readability
**Bad**:
```go
const (
    SmallCPU  = 1
    MediumCPU = 4
    LargeCPU  = 8
)

instance := NewInstanceType(SmallCPU)  // Reader must jump to find value
```

**Better**:
```go
instance := NewInstanceType(1)  // Explicit, clear intent
```

For test code, explicit values are often clearer than named constants.

### Rule of Thumb

| Usage Pattern | Location | Example |
|---|---|---|
| **1-3 tests** | Inline literal | `resource.MustParse("8")` |
| **One package** | Package-local const | `test/e2e/finder_test.go` |
| **Many packages** | `internal/testutil` | `DefaultClusterName` |
| **Whole objects** | Builder/factory | `NewInstanceType()` |
| **Reduces readability** | Delete it | Don't create it |

### Project Structure

```
internal/testutil/
├── builders.go          # Fluent builders: InstanceTypeBuilder, NodeClassBuilder
├── defaults.go          # Only truly global defaults (DefaultClusterName, etc.)
├── provider.go          # Provider helpers (MockKwokProvider, etc.)
├── nodeclass.go         # NodeClass factories and defaults
├── mocks.go             # Mock implementations
└── instancefixture/     # Instance builders (isolated to prevent import cycles)
    └── instance.go      # InstanceBuilder, PoweredOnInstance, PoweredOffInstance

test/e2e/
├── finder_test.go       # Finder test constants (vcsim inventory setup)
├── operator_test.go     # Operator test constants
├── instance_test.go     # Instance provider constants
├── session_test.go      # Session test constants
└── cloudprovider_test.go # CloudProvider e2e tests
```

### Why This Approach?

1. **Maintainability**: Constants live where they're used → easier to update
2. **Readability**: Developers see test setup in context, not scattered across files
3. **Scalability**: Builders grow gracefully; adding new fields just adds a new method
4. **Import safety**: No centralized file means fewer circular dependency issues
5. **Flexibility**: Builders enable composition and overrides; constants don't

### Current Implementation

✅ **Constants moved to test files**:
- `test/e2e/finder_test.go` — vcsim inventory constants
- `test/e2e/operator_test.go` — operator initialization constants
- `test/e2e/session_test.go` — session test constants
- `test/e2e/instance_test.go` — instance provider constants

✅ **Builders in testutil**:
- `InstanceTypeBuilder` — fluent API for corecloudprovider.InstanceType
- `NodeClassBuilder` — fluent API for v1alpha1.VsphereNodeClass
- `NodeClaimBuilder` — fluent API for karpv1.NodeClaim
- `NodeClassInstanceTypeBuilder` — fluent API for v1alpha1.InstanceType

✅ **Isolated builders**:
- `instancefixture.InstanceBuilder` — fluent API for pkg/providers/instance.Instance
  - Separated to prevent import cycles with instance_test.go
  - Used by cloudprovider_test.go (external package)

## Writing Tests Following This Strategy

### Example: Instance Type Test
```go
func TestSelectInstanceType_SelectsByPrice(t *testing.T) {
    // Use builder directly, no constants
    cheapest := testutil.NewInstanceType().
        WithName("c-2x").
        WithCPU("1").
        WithMemory("2Gi").
        WithPrice(0.27).
        Build()
    
    expensive := testutil.NewInstanceType().
        WithName("m-2x").
        WithCPU("1").
        WithMemory("8Gi").
        WithPrice(0.42).
        Build()
    
    selected := selectInstanceType([]InstanceType{expensive, cheapest})
    assert.Equal(t, "c-2x", selected.Name)
}
```

### Example: vSphere Test
```go
// test/e2e/finder_test.go
func TestFindDatacenter(t *testing.T) {
    p, ctx := setupProvider(t, FinderStandaloneHostPrefix)
    
    dc, err := p.FindClient.Datacenter(ctx, FinderDatacenterName)
    require.NoError(t, err)
    assert.NotNil(t, dc)
}
```

## When to Break These Rules

- **Shared numerical limits**: If multiple tests check "max 10 pods", that's fine as a named constant
- **Environment values**: CI/CD credentials might be centralized (though prefer env vars)
- **API enums**: If testing against a standard enum, the constant is the enum itself

**Key question**: Does moving this constant to its test file make the test **harder** to understand? If no, move it.
