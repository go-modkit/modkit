# D7: Validation Guide

**Status:** 🔴 Not started  
**Type:** New guide  
**NestJS Equivalent:** Pipes

---

## Goal

Document Go‑idiomatic validation and transformation patterns as the modkit equivalent of Nest pipes.

## Why Different from NestJS

NestJS pipes are framework hooks that transform/validate before handlers. In Go, validation is explicit in handlers using standard library or validation libraries. This is more verbose but debuggable.

## Files to Create/Modify

- Create: `docs/guides/validation.md`
- Modify: `README.md` (add link)

---

## Task 1: Create validation guide

**Files:**
- Create: `docs/guides/validation.md`

### Step 1: Draft the guide

Include:

1. **JSON decode + validation flow**
2. **Example with `json.Decoder` + `DisallowUnknownFields`**
3. **Optional mention of validator libraries** (no core dependency)
4. **Path parameter validation**

Suggested structure:

```markdown
# Validation

modkit uses explicit validation in handlers rather than framework-level pipes.

## JSON Request Validation

```go
func (c *Controller) CreateUser(w http.ResponseWriter, r *http.Request) {
    var req CreateUserRequest
    
    dec := json.NewDecoder(r.Body)
    dec.DisallowUnknownFields()
    if err := dec.Decode(&req); err != nil {
        httpapi.WriteProblem(w, http.StatusBadRequest, "Invalid JSON")
        return
    }
    
    if err := validate(req); err != nil {
        httpapi.WriteProblem(w, http.StatusBadRequest, err.Error())
        return
    }
    
    // proceed with valid request
}
```

## Manual Validation

```go
func validate(req CreateUserRequest) error {
    if req.Name == "" {
        return errors.New("name is required")
    }
    if !strings.Contains(req.Email, "@") {
        return errors.New("invalid email")
    }
    return nil
}
```

## Using a Validation Library

For complex validation, consider `go-playground/validator`:

```go
import "github.com/go-playground/validator/v10"

var validate = validator.New()

type CreateUserRequest struct {
    Name  string `json:"name" validate:"required"`
    Email string `json:"email" validate:"required,email"`
}

func (c *Controller) CreateUser(w http.ResponseWriter, r *http.Request) {
    var req CreateUserRequest
    // ... decode ...
    if err := validate.Struct(req); err != nil {
        httpapi.WriteProblem(w, http.StatusBadRequest, err.Error())
        return
    }
}
```

## Path Parameter Validation

```go
func (c *Controller) GetUser(w http.ResponseWriter, r *http.Request) {
    idStr := r.PathValue("id")
    id, err := strconv.ParseInt(idStr, 10, 64)
    if err != nil {
        httpapi.WriteProblem(w, http.StatusBadRequest, "Invalid user ID")
        return
    }
    // proceed with valid id
}
```
```

### Step 2: Commit

```bash
git add docs/guides/validation.md
git commit -m "docs: add validation guide"
```

---

## Task 2: Link guide from README

**Files:**
- Modify: `README.md`

### Step 1: Add validation guide to the Guides list

### Step 2: Commit

```bash
git add README.md
git commit -m "docs: link validation guide"
```
