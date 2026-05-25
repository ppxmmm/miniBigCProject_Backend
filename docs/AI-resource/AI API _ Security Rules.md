
# **🔐 AI API & Security Rules – REST & Validation**

**Purpose:** This file governs the external interface contract, security protocols, and data validation standards.

**Scope:** REST Design, HTTP Status Codes, Input Validation, and Security Best Practices.

## **1\. 🌐 REST API Standards**

### **1.1 Resource Design**

* **Naming:** Use **plural nouns** for resources. Use lower-case.  
* **Versioning:** All APIs must be versioned.  
  * ✅ GET /api/v1/orders  
  * ✅ POST /api/v1/users/{userID}/addresses  
  * ❌ GET /api/getOrders (No verbs in paths)  
  * ❌ POST /api/create\_user (Use standard POST to /users)

### **1.2 HTTP Methods**

| Method | Usage | Idempotent? |
| :---- | :---- | :---- |
| **GET** | Read resources. Never modify state. | Yes |
| **POST** | Create new resources or complex non-idempotent actions. | No |
| **PUT** | **Full** update. Replaces the entire resource. | Yes |
| **PATCH** | **Partial** update. Modifies specific fields. | No |
| **DELETE** | Remove a resource. | Yes |

## **2\. 🚦 Status Code Contract**

You **MUST** return the correct HTTP status code. Do not return 200 OK with an error message in the body.

| Code | Meaning | When to Use |
| :---- | :---- | :---- |
| **200** | OK | Successful GET, PUT, or PATCH. |
| **201** | Created | Successful POST (Creation). **MUST** return resource or ID. |
| **204** | No Content | Successful DELETE. (No body). |
| **400** | Bad Request | Validation failure, malformed JSON. |
| **401** | Unauthorized | Missing or invalid authentication token. |
| **403** | Forbidden | Valid token, but user lacks permission (RBAC). |
| **404** | Not Found | Resource does not exist. |
| **409** | Conflict | Domain rule violation (e.g., "Email already exists"). |
| **422** | Unprocessable | Semantic errors (valid JSON types, but invalid logic). |
| **500** | Internal Error | Unhandled exceptions, DB connection failures. |

## **3\. 🛡️ Input Validation**

### **3.1 Controller Responsibility**

Validation happens in the **Controller** layer before calling the Use Case.

1. **Syntactic Validation:** (Is it valid JSON? Are required fields present?)  
   * Handle json.Unmarshal errors strictly.  
   * Return 400 Bad Request.  
2. **Semantic Validation:** (Is the email format valid? Is age \> 18?)  
   * Use a validator library (e.g., go-playground/validator) on the Input DTO.

### **3.2 Code Example**

type CreateUserRequest struct {  
    Email string \`json:"email" validate:"required,email"\`  
    Age   int    \`json:"age"   validate:"gte=18"\`  
}

func (h \*UserHandler) Create(ctx \*fasthttp.RequestCtx) {  
    var req CreateUserRequest  
      
    // 1\. Parse  
    if err := json.Unmarshal(ctx.PostBody(), \&req); err \!= nil {  
        ctx.SetStatusCode(fasthttp.StatusBadRequest)  
        return  
    }

    // 2\. Validate  
    if err := h.validator.Struct(req); err \!= nil {  
        ctx.SetStatusCode(fasthttp.StatusBadRequest)  
        // Return structured validation error details  
        return  
    }

    // 3\. Proceed  
    h.useCase.Create(ctx, \&req)  
}

## **4\. 🔒 Security Best Practices**

### **4.1 Authentication & Context**

* **Middleware:** Authentication (JWT parsing) must happen in middleware.  
* **Context Injection:** valid User IDs / Roles must be injected into ctx.UserValue.  
* **Usage:** Controllers extract user info from context, **NEVER** from the request body.

// Middleware sets this:  
ctx.SetUserValue("userID", "123")

// Controller reads this:  
userID := ctx.UserValue("userID").(string)

### **4.2 Logging Security (PII)**

**STRICT RULE:** Never log Sensitive Data.

* ❌ Passwords / API Keys  
* ❌ Auth Tokens (JWTs)  
* ❌ Credit Card Numbers  
* ❌ Personally Identifiable Information (Names, Emails) \- *unless masked*

### **4.3 Headers**

* **Content-Type:** Always set ctx.SetContentType("application/json") for API responses.  
* **CORS:** Configure fasthttp CORS handler if accessed from browsers.

## **5\. 📦 Response Envelope**

Unless otherwise specified, API responses should be consistent.

**Success:**

{  
  "data": { ...resource... }  
}

**Error:**

{  
  "error": {  
    "code": "USER\_NOT\_FOUND",  
    "message": "The user with ID 123 does not exist.",  
    "details": \[\]  
  }  
}  
