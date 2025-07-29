# .gitignore File Documentation

## 📋 Overview

The `.gitignore` file in this project has been optimized for Go language projects and Hertz MVC framework to effectively filter out files that don't need version control.

## 🚫 Ignored File Types

### Go Language Related
- `*.exe, *.dll, *.so, *.dylib` - Build artifacts
- `*.test, *.out` - Test files
- `go.work, go.work.sum` - Go workspace files
- `vendor/` - Dependency management directory
- `*.cover, *.prof` - Coverage and profiling files

### Framework Related
- `logs/` - Log directory and all log files
- `*.env` - Environment configuration files
- `uploads/, cache/, sessions/` - Runtime directories
- `*.db, *.sqlite` - Database files

### Development Tools
- `.vscode/, .idea/` - IDE configurations
- `*.swp, *.swo, *~` - Editor temporary files

### System Files
- `.DS_Store` - macOS system files
- `Thumbs.db` - Windows thumbnails  
- `._*` - macOS resource fork files

### Project Specific
- `example_app, log_demo` - Example application binaries
- `benchmark_results/, pprof_data/` - Performance test results

## ✅ Verify .gitignore Effect

### Check ignored files
```bash
# Check if specific files are ignored
git check-ignore -v logs/ example_app

# View all ignored files
git status --ignored
```

### View current project status
```bash
# View working directory status
git status

# View concise status
git status --porcelain
```

## 🛠️ Custom Configuration

If you need to add project-specific ignore rules, please add them at the end of the "Project Specific" section in `.gitignore`:

```gitignore
# ============= Project Specific =============

# Your custom rules
my_custom_file.txt
my_custom_dir/
```

## 📚 Common Scenarios

### 1. Development Environment Configuration
```gitignore
# Development environment specific config
.env.local
config.dev.yaml
```

### 2. Build Artifacts
```gitignore
# Build output
bin/
dist/
build/
```

### 3. Temporary Files
```gitignore
# Temporary and cache files
tmp/
*.tmp
*.cache
```

## ⚠️ Important Notes

1. **Configuration files**: Sensitive production config files should be ignored, but keep example config files
2. **Log files**: All log files are ignored to avoid committing large amounts of log data
3. **Build artifacts**: Binary files and build artifacts should not be committed to version control
4. **IDE configuration**: Personal development tool configurations are ignored to avoid affecting team members

## 🔍 Troubleshooting

### File unexpectedly ignored
```bash
# Check why a file is being ignored
git check-ignore -v path/to/file

# Force add ignored file
git add -f path/to/file
```

### File not being ignored
```bash
# Check .gitignore syntax
git check-ignore --no-index path/to/file

# Re-apply .gitignore
git rm -r --cached .
git add .
```

## 📖 Best Practices

1. **Regular updates**: Update .gitignore rules as the project evolves
2. **Team collaboration**: Ensure all team members understand the ignore rules
3. **Layered management**: Add additional .gitignore files in subdirectories if needed
4. **Test verification**: Verify that important files are correctly ignored before committing

## 🚀 Quick Test

Run the following commands to test if .gitignore is working properly:

```bash
# Build example application
go build -o example_app ./example/main.go

# Check if build artifact is ignored
git status

# Should not see example_app file
```

## 📝 File Categories

The .gitignore file is organized into the following categories:

1. **Go Language Project** - Go-specific files and artifacts
2. **Framework Related Files** - Runtime files, logs, configs
3. **Development Tools** - IDE and editor configurations
4. **System Files** - OS-specific files
5. **Version Control** - Git-related files
6. **Testing Related** - Test outputs and benchmarks
7. **Deployment Related** - Build artifacts and archives
8. **Certificates and Keys** - Security-related files
9. **Cloud Service Configuration** - Cloud provider configs
10. **CI/CD** - Continuous integration files
11. **Project Specific** - Custom project files

Each category is clearly marked with comments for easy maintenance and understanding.