# Go Project Health Inspector 🩺

> "Think of this tool as a doctor, not a compiler. It doesn’t say 'this is invalid Go'. It says 'this will hurt you in production'."

**goHealth** is an opinionated static analysis tool designed to catch **production risks**, **concurrency traps**, and **resource leaks** that standard linters often miss. It prioritizes *behavioral correctness* over style.

## 🚀 Features

### The "Doctor" Philosophy
*   **Production-Minded**: Focuses on what kills services (blocking loops, leaks, injection).
*   **Educational**: Doesn't just flag lines; explains *why* the code is dangerous.
*   **Zero Config**: No flags to disable checks. It's a consultant you hire, not a tool you tweak.

### Checks Implemented

#### 1. Concurrency & Performance
*   **Blocking Sleep**: Detects `time.Sleep` inside loops (blocks the goroutine/thread).
*   **Context Misuse**: Detects `context.Context` stored in structs (risk of memory leaks/stale contexts).
*   **HTTP Client Timeouts**: Detects `http.Get`, `http.Post`, or `http.Client{}` without timeouts (causes hangs and crashes).
*   **Broken Context Chain**: Detects calls with `context.TODO()` or `context.Background()` when a valid context is available (breaks cancellation and tracing).

#### 2. Resource Management
*   **HTTP Body Leaks**: Detects `http.Response` bodies that aren't closed (file descriptor leaks).
*   **DB Connection Leaks**: Detects `sql.Rows` that aren't closed (connection pool exhaustion).
*   **Zombie Transactions**: Detects database transactions started without deferred rollback (locks and connection leaks).

#### 3. Security
*   **SQL Injection**: Detects `fmt.Sprintf` or string concatenation used to build SQL queries.
*   **Weak Randomness**: Detects `math/rand` usage for token/password generation (predictable RNG).
*   **Exit in Libraries**: Detects `os.Exit`, `log.Fatal`, or `panic` calls in non-main packages (prevents graceful error handling).

#### 4. Performance & Reliability
*   **Memory DOS Protection**: Detects `io.ReadAll` without `io.LimitReader` (prevents OOM from large payloads).
*   **Silenced Critical Errors**: Detects ignoring errors from `json.Marshal`, `db.Exec`, etc. (hides failures).
*   **Slice Append Races**: Detects appending to slices in goroutines without synchronization (data races).
*   **Time Comparison Bugs**: Detects `time.Time == time.Time` instead of `.Equal()` (unreliable equality).
*   **Empty Spin Loops**: Detects CPU-burning `for { select { default: } }` loops (busy-wait).

## 📦 Installation & Usage

### Inspect a single package
```bash
go run main.go check .
```

### Inspect the entire project (recursive)
```bash
go run main.go check ./...
```

### Example Output
```text
🚨 [CRITICAL] Potential SQL Injection risk detected.
   📍 Location: /path/to/app/db.go:42
   📝 Code: db.Query(fmt.Sprintf(...))

   🎓 Why this matters:
   Building SQL queries using fmt.Sprintf allows attackers to inject malicious SQL. Use parameterized queries (e.g., $1, ?) and pass arguments separately to db.Query().
```

## 🛠 Project Structure

*   `internal/doctor`: Core interface definitions (`Check`, `Diagnosis`).
*   `internal/checks`: The logic for individual health checks.
*   `internal/report`: Console output formatting.
*   `main.go`: CLI entry point.

---
*Built with ❤️ for better Go services.*
