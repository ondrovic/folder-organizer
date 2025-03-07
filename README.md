# Folder Organizer

A powerful CLI tool that organizes files in directories based on file extensions and a customizable configuration.

## Features

- **Flexible Organization**: Sort files into directories based on file extensions
- **Hierarchical Categories**: Support for nested category structures
- **Extension-based Sub-folders**: Files are organized into extension-specific sub-folders
- **Multi-threaded**: Efficiently process files using concurrent operations
- **Progress Display**: Real-time progress tracking during organization
- **Empty Directory Cleanup**: Option to remove empty directories after organization
- **Cross-platform**: Works on Windows, macOS, and Linux

## Installation

### From Source

```bash
# Clone the repository
git clone https://github.com/ondrovic/folder-organizer.git
cd folder-organizer

# Build the application
make build
```

### Using Go Install

```bash
go install github.com/ondrovic/folder-organizer@latest
```

## Usage

Basic usage:

```bash
folder-organizer organize config.json /path/to/folder
```

With options:

```bash
folder-organizer organize --workers=8 --recursive=true --progress --cleanup config.json /path/to/folder
```

### Command Options

- `--workers, -w`: Number of worker goroutines (default: 4)
- `--recursive, -r`: Process subdirectories recursively (default: true)
- `--progress, -p`: Show progress during organization (default: false)
- `--cleanup, -c`: Remove empty directories after organization (default: false)

## Configuration

The folder organizer uses a JSON configuration file to define how files should be organized. The configuration file uses a hierarchical structure to define categories and subcategories:

### Basic Configuration

```json
{
  "categories": {
    "images": [".jpg", ".png", ".gif"],
    "documents": [".pdf", ".docx", ".txt"],
    "videos": [".mp4", ".avi", ".mkv"]
  }
}
```

### Advanced Configuration with Nested Categories

```json
{
  "categories": {
    "images": [".jpg", ".png", ".gif", ".bmp", ".jpeg", ".svg", ".webp"],
    "documents": [
      {
        "excel": [".xls", ".xlsx", ".xlsm", ".xltx", ".xltm"],
        "word": [".doc", ".docx", ".docm", ".dotx", ".dotm"],
        "text": [".txt"]
      },
      [".pdf"]
    ],
    "design": [
      {
        "fusion": [".f3d"]
      },
      {
        "printing": [".stl", ".step", ".obj", ".3mf"]
      }
    ]
  }
}
```

## File Organization Structure

Files are organized into the following structure:

```
source-dir/
├── category/
│   └── extension/
│       └── filename.ext
├── nested-category/
│   └── subcategory/
│       └── extension/
│           └── filename.ext
```

For example, with the advanced configuration above:

- `photo.jpg` → `/source-dir/images/jpg/photo.jpg`
- `document.docx` → `/source-dir/documents/word/docx/document.docx`
- `model.f3d` → `/source-dir/design/fusion/f3d/model.f3d`

## Development

### Project Structure

```
 .
├──  .github/               # GitHub configuration
│   ├──  ISSUE_TEMPLATE/    # Issue templates for GitHub
│   │   ├──  bug_report.md
│   │   ├──  config.template.yml
│   │   ├──  feature_request.md
│   │   └──  question.md
│   ├──  scripts/           # CI/CD scripts
│   │   └──  generate_config.sh
│   └──  workflows/         # GitHub Actions workflows
│       ├──  generate-config.yml
│       ├──  releaser.yml
│       └──  testing.yml
├──  .gitignore             # Git ignore file
├──  .goreleaser.yaml       # GoReleaser configuration
├──  cmd/                   # Command-line interface
│   └──  cli/               # CLI commands
│       ├──  organize.go    # Organize command implementation
│       ├──  root.go        # Root command definition
│       └──  version.go     # Version command implementation
├──  folder-organizer.go    # Main application entry point
├──  go.mod                 # Go module file
├──  go.sum                 # Go dependencies checksums
├──  internal/              # Internal packages
│   ├──  types/             # Type definitions
│   │   └──  types.go       # Core types for the application
│   └──  utils/             # Utility functions
│       ├──  cleaner.go     # Empty directory cleanup
│       ├──  organize.go    # File organization logic
│       └──  progress.go    # Progress tracking utilities
├──  LICENSE                # License information
├──  Makefile               # Build automation
└──  README.md              # Project documentation
```

### Building from Source

```bash
# Build for current platform
make build

# Build for all platforms
make build-all

# Run tests
make test
```

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.