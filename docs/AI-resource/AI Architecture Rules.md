# **🏛️ AI Architecture Rules – Go Clean Architecture**

**Purpose:** This file governs the project structure, layer responsibilities, and dependency rules.

**Scope:** Directory structure, package naming, sequence diagram interpretation, and strict Clean Architecture enforcement.

## **1\. 📋 Sequence Diagrams (THE SOURCE OF TRUTH)**

**CRITICAL RULE:** Before scaffolding any code, you MUST map the Sequence Diagram participants to the folder structure.

### **1.1 Interpretation Rules**

1. **Folder Path Mapping:** The text inside the parentheses of a participant definition is the **MANDATORY** file location.  
   participant "PromotionRuleService\\n(application/service)" as Service

   * **Class Name:** PromotionRuleService  
   * **Location:** application/service (Do not change or simplify this path)  
2. **Flow Enforcement:** Implement the interaction flow exactly as shown.  
   * If Controller \-\> Service \-\> Repository, do NOT skip layers.  
   * If Service \-\> Service: transform(), the transformation logic belongs in the Service.  
3. **Method Signatures:** Use the exact method names and return types defined in the diagram.

### **1.2 Path Mapping Reference**

| Participant Type | Implementation Name | Required Folder Path |
| :---- | :---- | :---- |
| **Handler** | {Entity}Handler | internal/controller/http/ |
| **UseCase** | {Entity}UseCase | internal/usecase/ |
| **Repository** | {Entity}Repository | internal/repository/postgres/ |
| **Entity** | {Entity} | internal/entity/ |

## **2\. 📂 Project Structure**

Maintain this structure rigidly. Do not invent new top-level directories.

.  
├── cmd/api/main.go                        \# Entry point  
├── internal/  
│   ├── entity/                            \# INNERMOST: Core domain models (No tags, no deps)  
│   ├── usecase/                           \# APPLICATION: Business logic & Interface Definitions  
│   │   ├── dto/                           \# Data Transfer Objects (Input/Output)  
│   │   └── {entity}\_usecase.go            \# Implementation  
│   ├── controller/http/                   \# ADAPTER: HTTP Handlers (fasthttp) & Router  
│   ├── repository/postgres/               \# ADAPTER: Database implementation (pgx)  
│   ├── infrastructure/                    \# OUTERMOST: Config, External Clients, Drivers  
│   └── presenter/                         \# ADAPTER: Response formatting  
└── pkg/                                   \# Public shared libraries

## **3\. 🛡️ The Dependency Rule**

**"Source code dependencies must point INWARD only."**

1. **Entity Layer:** Knows NOTHING about any other layer.  
2. **Use Case Layer:** Knows only about **Entities**.  
   * *Note:* Repository interfaces are defined HERE (Inner Layer), but implemented in the Repository Layer (Outer Layer).  
3. **Adapters (Controllers/Repos):** Know about **Use Cases** and **Entities**.  
4. **Infrastructure:** Knows about everything.

**❌ Strict Prohibitions:**

* **NEVER** import controller, repository, or infrastructure into entity or usecase.  
* **NEVER** use struct tags (JSON, DB) in the entity layer.  
* **NEVER** return Entities directly from Controllers (Use Presenters or DTOs).

## **4\. 🏷️ Naming & Layer Responsibilities**

| Layer | File Naming | Responsibility |
| :---- | :---- | :---- |
| **Entity** | order.go | **Core Rules.** Plain Go structs. Domain errors. No external libs. |
| **Use Case** | order\_usecase.go | **Orchestration.** Defines interfaces (Repository/Gateway). Application logic. |
| **Repository** | order\_repository.go | **Data Access.** Implements Use Case interfaces. Maps DB Models ↔ Entities. |
| **Controller** | order\_handler.go | **Transport.** Validates HTTP input. Calls Use Cases. Handles Status Codes. |
| **DTO** | order\_input.go | **Transport Data.** Pure data carriers between Controller and Use Case. |

## **5\. 🧩 Dependency Inversion Principle**

To ensure the Use Case layer remains independent of the database, strictly follow this interface pattern:

1. **Define Interface in Use Case:**  
   The Use Case defines the contract it needs.  
   * Location: internal/usecase/repository.go (or specific use case file)  
   * Code: type OrderRepository interface { ... }  
2. **Implement in Repository:**  
   The Repository adheres to that contract.  
   * Location: internal/repository/postgres/order\_repository.go  
   * Code: func (r \*OrderRepository) Save(...) { ... }

## **6\. 🔄 Standard Data Flow**

1. **HTTP Request** → Controller (Parses JSON to Input DTO)  
2. **Controller** → Use Case (Passes Input DTO)  
3. **Use Case** → Entity (Validates business rules)  
4. **Use Case** → Repository Interface (Requests data persistence)  
5. **Repository Impl** → Database (Maps Entity to DB Model & executes query)  
6. **Database** → Repository Impl → Entity (Maps back to Entity)  
7. **Use Case** → Output DTO (Transforms Entity for response)  
8. **Controller** → **HTTP Response** (JSON)