# **🧪 AI Testing & QA Rules – Quality Assurance**

**Purpose:** This file governs the testing strategy, build requirements, and the definition of "Done".

**Scope:** Unit Tests, Integration Tests, Linting, and Compilation Checks.

## **1\. 🟢 The Green Build Requirement (MANDATORY)**

**Critical Rule:** A task is **NEVER** complete until the build is green.

### **1.1 Verification Checklist**

After every code change, you **MUST** run the following sequence. If any step fails, fix it immediately.

1. **Compile:** go build ./...  
   * ✅ Must pass with NO errors.  
2. **Test:** go test ./...  
   * ✅ All tests must pass.  
3. **Vet:** go vet ./...  
   * ✅ Must catch common Go mistakes.  
4. **Lint:** golangci-lint run  
   * ✅ Code must adhere to style guides.  
5. **Dependencies:** go mod tidy  
   * ✅ go.mod and go.sum must be clean.

## **2\. 🧪 Testing Strategy by Layer**

### **2.1 Domain Layer (Unit Tests)**

* **Scope:** internal/entity/  
* **Technique:** Pure unit tests. **NO Mocks**. **NO DB**.  
* **Goal:** Verify business rules and value object validation.  
* **File:** entity/order\_test.go

func TestOrder\_Validate(t \*testing.T) {  
    tests := \[\]struct {  
        name    string  
        total   float64  
        wantErr bool  
    }{  
        {"Valid", 100.0, false},  
        {"Invalid", \-10.0, true},  
    }  
    for \_, tt := range tests {  
        o := Order{Total: tt.total}  
        if err := o.Validate(); (err \!= nil) \!= tt.wantErr {  
            t.Errorf("Validate() error \= %v, wantErr %v", err, tt.wantErr)  
        }  
    }  
}

### **2.2 Use Case Layer (Application Tests)**

* **Scope:** internal/usecase/  
* **Technique:** Unit tests with **Mocked Interfaces**.  
* **Goal:** Verify orchestration logic (Did it call the repo? Did it handle the error?).  
* **Mocks:** Use stretchr/testify/mock or manual mocks.

// Mocking the Repository Interface defined in Use Case  
type MockRepo struct { mock.Mock }  
func (m \*MockRepo) Save(ctx context.Context, o \*entity.Order) error {  
    args := m.Called(ctx, o)  
    return args.Error(0)  
}

func TestCreateOrder(t \*testing.T) {  
    // Setup  
    mockRepo := new(MockRepo)  
    uc := NewOrderUseCase(mockRepo)  
      
    // Expectation  
    mockRepo.On("Save", mock.Anything, mock.Anything).Return(nil)  
      
    // Execution  
    \_, err := uc.Create(context.Background(), \&dto.CreateOrderInput{})  
      
    // Assertion  
    assert.NoError(t, err)  
    mockRepo.AssertExpectations(t)  
}

### **2.3 Integration Tests (System Tests)**

* **Scope:** End-to-End flow (Controller → DB).  
* **Technique:** Use testcontainers-go to spin up a real PostgreSQL Docker container.  
* **Tags:** Must use //go:build integration.  
* **Goal:** Verify SQL queries, transaction handling, and HTTP contract.

//go:build integration

func TestOrderFlow(t \*testing.T) {  
    // 1\. Start Postgres Container  
    // 2\. Migrate DB  
    // 3\. Send HTTP Request to Handler  
    // 4\. Verify DB state  
}

## **3\. 🚫 Testing Anti-Patterns**

* **❌ No Logic in Tests:** Do not put complex branching logic in test code.  
* **❌ No Sleep:** Do not use time.Sleep() to wait for async events (use channels/waitgroups).  
* **❌ No Global State:** Tests must be parallel-safe (t.Parallel()).  
* **❌ Mocking Entities:** Never mock an Entity struct. Use the real struct.  
* **❌ Testing Private Methods:** Only test the public API (exported methods).

## **4\. 📝 Error Debugging Protocol**

If the build fails, follow this protocol:

1. **Read** the exact error message from the compiler/linter.  
2. **Locate** the file and line number.  
3. **Fix** the root cause (e.g., missing import, type mismatch).  
4. **Re-run** the full checklist (Section 1.1).  
5. **NEVER** comment out failing tests to make the build pass.