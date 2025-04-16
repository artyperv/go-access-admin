# Go Access Admin

A web-based administration tool for managing access to protected resources using `.htpasswd` files. This application provides a user-friendly interface to create, manage, and monitor access credentials with automatic expiration.

## Features

- Web-based interface for managing access credentials
- Support for multiple `.htpasswd` files
- Automatic password generation
- Configurable access duration (1 hour, 1 day, 1 week)
- Automatic cleanup of expired accesses
- URL template support for generating access links

## Prerequisites

- Go 1.24.2 or later

## Installation

1. Clone the repository:
```bash
git clone https://github.com/yourusername/go-access-admin.git
cd go-access-admin
```

2. Create a configuration file:
```bash
cp internal/config/config.yaml.example internal/config/config.yaml
```

3. Edit the configuration file to match your environment:
```yaml
app:
  debug: false
  sync_htpasswd: true  # deletes all records from .htpasswd and writes db users in it on startup
  clean_accesses_interval: 1  # in minutes

htpasswd_paths:
  - name: "example_service"
    path: "/path/to/.htpasswd"
    url_template: "https://{user}:{password}@site.ru/protected"
    admins:
      - username: "admin"
        password: "secure_password"

# Global admins
admins:
  - username: "admin"
    password: "secure_password"
```

4. Build and run the application:
```bash
go build
./go-access-admin
```

The application will be available at `http://localhost:8080`.

## Usage

1. Access the web interface at `http://localhost:8080`
2. Select the desired `.htpasswd` file from the dropdown
3. Create new access credentials:
   - Enter username (required)
   - Enter password (optional - will be auto-generated if left empty)
   - Select access duration
   - Click "Add"
4. Manage existing accesses:
   - View access details
   - Copy access links
   - Delete expired or unnecessary accesses

## Project Structure

```
go-access-admin/
├── internal/
│   ├── access/         # .htpasswd file operations
│   ├── config/         # Configuration management
│   ├── handler/        # HTTP request handlers
│   ├── scheduler/      # Background tasks
│   └── storage/        # Database operations
├── web/
│   ├── static/         # Static assets
│   └── templates/      # HTML templates
└── main.go            # Application entry point
```
