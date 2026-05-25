# **🛠️ AI Implementation Rules – Go Coding Standards**

**Purpose:** This file governs **how** code is written. It defines the mandatory libraries, syntax patterns, and implementation details for each layer.

**Scope:** Go 1.22+, fasthttp, pgx/v5, Logging, Error Handling, and Context usage.

## **1\. 🧱 Mandatory Tech Stack**

You **MUST** use these specific libraries. Do not substitute them with standard library equivalents (e.g., do not use net/http or database/sql).

| Component | Library/Version | Strict Rule |
| :---- | :---- | :---- |
| **Language** | Go 1.22+ | Use modern Go idioms. |
| **HTTP Server** | github.com/valyala/fasthttp | **NEVER** use net/http. |
| **Database** | github.com/jackc/pgx/v5 | Use pgxpool. **NEVER** use lib/pq. |
| **Logging** | gitlab.com/bigc-dtt/gomodule/log | Use this specific internal wrapper. |
| **Router** | fasthttp/router (or compatible) | Compatible with fasthttp context. |

## **2\. ⚡ Core Coding Patterns**

### **2.1 Context Propagation**

* **Rule:** context.Context must be the **first argument** of every function in the Controller, UseCase, and Repository layers.  
* **Reason:** Required for tracing, cancellation, and logging.

### **2.2 Error Handling**

* **Entity:** Return domain errors (e.g., ErrInsufficientFunds).  
* **Use Case:** Propagate errors or wrap them if necessary.  
* **Controller:** Map errors to HTTP Status Codes. **NEVER** panic.

### **2.3 Structured Logging**

* **Import:** gitlab.com/bigc-dtt/gomodule/log  
* **Usage:** Log at **Controller** (entry/exit) and **Use Case** (business logic flow) boundaries.  
* **Syntax:**  
  logger.Info(ctx, "Order processed", log.Field("orderID", id))  
  logger.Error(ctx, "Failed to create order", log.Field("error", err))

## **3\. 📝 Layer-Specific Implementation**

### **3.1 Entity Layer (internal/entity/)**

* **Content:** Plain Go structs \+ Domain logic methods.  
* **Forbidden:** Struct tags (json:"...", db:"..."), external imports.

type Order struct {  
    ID    string  
    Total float64  
}

// Domain Logic  
func (o \*Order) Validate() error {  
    if o.Total \<= 0 {  
        return errors.New("total must be positive")  
    }  
    return nil  
}

### **3.2 Use Case Layer (internal/usecase/)**

* **Content:** Interfaces \+ Business Logic \+ DTO Conversion.  
* **Forbidden:** fasthttp, pgx imports.

// 1\. Define Interface (Contract)  
type OrderRepository interface {  
    Save(ctx context.Context, order \*entity.Order) error  
}

// 2\. Implementation  
func (uc \*OrderUseCase) Create(ctx context.Context, input \*dto.CreateOrderInput) (\*dto.OrderOutput, error) {  
    // Convert DTO \-\> Entity  
    order := input.ToEntity()  
      
    // Business Logic  
    if err := order.Validate(); err \!= nil {  
        return nil, err  
    }

    // Call Repository  
    if err := uc.repo.Save(ctx, order); err \!= nil {  
        return nil, err  
    }

    // Convert Entity \-\> DTO  
    return dto.FromEntity(order), nil  
}

### **3.3 HTTP Controller (internal/controller/http/)**

* **Content:** Request parsing, validation, Use Case invocation, Response formatting.  
* **Library:** fasthttp.

func (h \*OrderHandler) CreateOrder(ctx \*fasthttp.RequestCtx) {  
    // 1\. Parse Body  
    var input dto.CreateOrderInput  
    if err := json.Unmarshal(ctx.PostBody(), \&input); err \!= nil {  
        ctx.SetStatusCode(fasthttp.StatusBadRequest)  
        return  
    }

    // 2\. Call Use Case (Convert fasthttp context to standard context)  
    res, err := h.useCase.Create(ctx, \&input)  
    if err \!= nil {  
        h.logger.Error(ctx, "creation failed", log.Field("err", err))  
        ctx.SetStatusCode(fasthttp.StatusInternalServerError) // Map error to status  
        return  
    }

    // 3\. Send Response  
    ctx.SetStatusCode(fasthttp.StatusCreated)  
    ctx.SetContentType("application/json")  
    json.NewEncoder(ctx).Encode(res)  
}

### **3.4 Repository (internal/repository/postgres/)**

* **Content:** pgx implementation, Entity ↔ DB Model mapping.  
* **Library:** pgx/v5.

type OrderRepository struct {  
    db \*pgxpool.Pool  
}

func (r \*OrderRepository) Save(ctx context.Context, order \*entity.Order) error {  
    // 1\. Map Entity \-\> DB Model  
    model := toModel(order)

    // 2\. Execute Query  
    \_, err := r.db.Exec(ctx, "INSERT INTO orders ...", model.ID, model.Total)  
    return err  
}

func (r \*OrderRepository) FindByID(ctx context.Context, id string) (\*entity.Order, error) {  
    var model OrderModel  
    // 3\. Query & Scan  
    err := r.db.QueryRow(ctx, "SELECT ...", id).Scan(\&model.ID, \&model.Total)  
    if err \!= nil {  
        return nil, err  
    }  
    // 4\. Map DB Model \-\> Entity  
    return toEntity(\&model), nil  
}

## **4\. 🚫 Strict Anti-Patterns**

1. **NO net/http:** Do not use the standard http.ResponseWriter or \*http.Request.  
2. **NO Tags in Entities:** Keep internal/entity pure. Use Mappers.  
3. **NO Logic in Controllers:** Controllers only parse requests and call Use Cases.  
4. **NO Logic in Repositories:** Repositories only Read/Write data.  
5. **NO Global State:** Use Dependency Injection (pass dependencies in New... functions).