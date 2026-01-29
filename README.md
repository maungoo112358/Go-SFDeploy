# SFDeploy

A Go-based CLI tool for automating hot deployment of Java extensions to SmartFox Server 2X.

## Overview

SFDeploy streamlines the development workflow for SmartFox Server 2X extensions by automating the entire deployment pipeline:

- **Build** - Compiles Java source files and packages them into JAR files
- **Deploy** - Copies JARs and configuration files to the SmartFox extensions directory
- **Restart** - Gracefully restarts the SmartFox server
- **Cleanup** - Removes temporary build artifacts

## Features

- Automatic Java 11 detection (JAVA_HOME, PATH, common installation directories)
- SmartFox Server directory validation
- Classpath auto-configuration from SmartFox lib directory
- Separate common library JAR packaging
- Process management for SmartFox server (port 9933)
- JSON configuration file deployment
- Detailed console output with progress indicators

## Requirements

- Go 1.24+ (for building from source)
- Java 11 (JDK required for javac and jar commands)
- SmartFox Server 2X installed
- Windows (primary platform; Linux/macOS support is partial)

## Installation

### From Source

```bash
git clone https://github.com/yourusername/SFDeploy.git
cd SFDeploy
go build -o sfdeploy
```

### Pre-built Binary

Download the latest release from the Releases page.

## Configuration

Create or edit `sfdeploy_config.json` in the same directory as the executable:

```json
{
  "java_path": "C:\\Program Files\\Java\\jdk-11\\bin",
  "source_dir": "C:\\Projects\\MyGame\\GameExtension",
  "target_dir": "C:\\SmartFoxServer_2X",
  "extension_folder": "MyExtension",
  "extension_file": "MyExtension.jar",
  "common_file": "CommonLib.jar",
  "common_folder": "common",
  "json_source_dir": "C:\\Projects\\MyGame\\GameExtension\\src\\config\\jsons",
  "deploy_json_files": ["GameConfig", "LevelData", "PlayerSettings"]
}
```

### Configuration Fields

| Field | Description |
|-------|-------------|
| `java_path` | Path to Java 11 bin directory (optional if Java is in PATH) |
| `source_dir` | Root directory of your Java extension project |
| `target_dir` | SmartFox Server 2X installation directory |
| `extension_folder` | Name of the extension folder within SmartFox extensions directory |
| `extension_file` | Output JAR filename for the main extension |
| `common_file` | Output JAR filename for shared/common library code |
| `common_folder` | Subfolder in src/ containing common library code |
| `json_source_dir` | Directory containing JSON configuration files to deploy |
| `deploy_json_files` | List of JSON filenames (without .json extension) to copy |

## Usage

Run the executable from the command line:

```bash
./sfdeploy
```

The tool will execute the following phases:

```
Phase 1: Directory Setup
  - Validates configuration file
  - Checks source and target directories
  - Verifies Java 11 installation

Phase 2: Building Project
  - Cleans old .class files
  - Compiles all Java source files
  - Creates common library JAR
  - Creates extension JAR

Phase 3: Deploying Project
  - Terminates processes on port 9933
  - Copies common JAR to SmartFox __lib__ folder
  - Copies extension JAR to SmartFox extensions folder
  - Deploys JSON configuration files

Phase 4: Restarting SmartFox Server
  - Launches SmartFox server with logging

Phase 5: Cleaning Up
  - Removes compiled .class files
  - Deletes temporary JARs from source directory
```

## Project Structure

```
SFDeploy/
├── main.go              # Entry point and workflow orchestration
├── config.go            # Configuration loading and validation
├── build.go             # Java compilation and JAR creation
├── deploy.go            # File deployment and cleanup
├── server.go            # SmartFox server management
├── utils.go             # Utility functions (Java detection, prompts)
├── sfdeploy_config.json # Configuration file
└── go.mod               # Go module definition
```

## Source Directory Requirements

Your Java project source directory must have:

- A `src/` subdirectory containing `.java` files
- Standard Java package structure

Example:

```
GameExtension/
├── src/
│   ├── common/              # Common library code (optional)
│   │   └── SharedUtils.java
│   └── com/
│       └── mycompany/
│           └── game/
│               ├── MainExtension.java
│               └── handlers/
│                   └── LoginHandler.java
└── lib/                     # Optional external dependencies
```

## SmartFox Server Requirements

The tool validates the target directory contains:

- `SFS2X/` directory
- `sfs2x.bat` launcher script
- `lib/` directory (used for classpath construction)

## Troubleshooting

### Java Not Found

The tool searches for Java 11 in:

1. JAVA_HOME environment variable
2. System PATH
3. Common Windows paths:
   - `C:\Program Files\Eclipse Adoptium\jdk-11*`
   - `C:\Program Files\Java\jdk-11*`
   - `C:\Program Files\OpenJDK\jdk-11*`

If not found, you'll be prompted to enter the path manually.

### Port 9933 Already in Use

The tool automatically terminates processes using port 9933 before deployment. If this fails, manually stop SmartFox Server before running the tool.

### Compilation Errors

Check the console output for javac error messages. Common issues:

- Missing dependencies in SmartFox `lib/` directory
- Syntax errors in Java source files
- Incompatible Java version

## License

This project is licensed under the MIT License. See the [LICENSE](LICENSE) file for details.

## Contributing

Contributions are welcome! To contribute:

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/your-feature`)
3. Commit your changes (`git commit -m "Add your feature"`)
4. Push to the branch (`git push origin feature/your-feature`)
5. Open a Pull Request

Please ensure your code follows the existing style and includes appropriate error handling.

## Author

Slint

## Support

If you encounter any issues or have questions, please open an issue on GitHub.
