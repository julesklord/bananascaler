# 🧹 [Code Health] Reduce memory allocations in profile lookups

## 🎯 What
Refactored `profileDB()` in `internal/hardware/profile.go` from a function returning a new map to a statically initialized package-level variable.

## 💡 Why
The previous implementation re-allocated a large map and multiple struct pointers every time `GetProfile()` was called. Converting this to a package-level variable initialization eliminates these repetitive allocations. This fulfills the request of addressing the issue without introducing an external JSON or YAML file that would have forced a major refactor of application initialization and fallback behavior. Additionally, `GetProfile` was updated to perform a shallow clone of the returned `UpscaleProfile` to ensure callers cannot accidentally mutate the global application state.

## ✅ Verification
1. Ensured all unit tests in the project (`go test ./...`) still pass.
2. Verified using `go run main.go detect` that CLI logic around retrieving and printing hardware tier presets works correctly.
3. Formatted with `go fmt ./...`.
4. Got the modified code successfully reviewed.

## ✨ Result
Lower memory overhead and garbage collection pressure when running queries against hardware profiles, while preserving strict global safety against state mutation.
