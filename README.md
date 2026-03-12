# Mac Cleaner (Go) 🧹

A blazing fast, safe, and beautiful terminal-based disk cleaner for macOS written in Go. 


## 🚀 Installation

Ensure you have [Go](https://go.dev/doc/install) installed.

### Quick Install (via `go install`)

```bash
go install github.com/mzavhorodnii/mac-cleaner-go/cmd/macclean@latest
```
*(Make sure your `$(go env GOPATH)/bin` is in your system `$PATH`.)*

## 💻 Usage

Simply run the app. By default, it will recursively scan your `/Users` directory for safe-to-delete targets:

```bash
macclean 
```

**Scan a specific directory:**
```bash
macclean ~/Projects
```

### Keyboard Controls:
* **`↑` or `k`**: Move up
* **`↓` or `j`**: Move down
* **`Enter`**: Clean the specially selected directory (or run "Clear All Items")
* **`/`**: Filter/search the list by name
* **`q` or `Ctrl+C`**: Quit gracefully

## 🛡️ What does it clean?
To ensure your system remains stable, `mac-cleaner-go` ignores standard user files and exclusively targets directories ending with the following names:
* `Caches`
* `Logs`
* `DerivedData` (Xcode build artifacts)
* `.Trash`
* `node_modules` (Heavy Javascript dependencies)

Empty directories (0 bytes) are automatically hidden from the UI to keep your workflow clean.
