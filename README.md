# gockapi

A CLI tool for setting up local mock servers using JSON configuration files.

---

## Features
- Spin up local mock servers for API development and testing
- Define endpoints, responses, and ports via simple JSON files
- Start all or individual mock services from a directory
- Hot-reload support for configuration changes

---

## Installation

### 1. Build and Install Locally

You can build and install `gockapi` locally without needing a release:

```bash
# Build the binary and place it in your Go bin directory
 go build -o ~/go/bin/gockapi ./cmd/gockapi

# Or use go install from the project root
 go install ./cmd/gockapi
```

### 2. Add to PATH

Make sure your Go bin directory (usually `~/go/bin`) is in your `PATH`:

```bash
export PATH=$PATH:~/go/bin
```
To make this change permanent, add the above line to your `~/.bashrc` or `~/.profile`.

---

## Usage

### Show Help
```bash
gockapi --help
```

---

## Example: Configure, Start, Use, and Stop a Mock Server

### 1. Create a Configuration File
Create a directory for your configs (e.g., `my-configs`) and add a file named `serviceA.json`:

```json
{
  "service_name": "serviceA",
  "port": 55001,
  "endpoints": {
    "GET /api/hello": {
      "status_code": 200,
      "headers": {
        "Content-Type": "application/json"
      },
      "body": {
        "message": "Hello from serviceA!"
      }
    }
  }
}
```

### 2. Start the Mock Server in the Background
```bash
gockapi --config-path ./my-configs start serviceA &
```
- This will start the mock server for `serviceA` in the background.
- You will see logs confirming the server is running.

### 3. Make a Request to Your Mock Server
```bash
curl http://localhost:55001/api/hello
```
- You should receive a JSON response:
  ```json
  {"message":"Hello from serviceA!"}
  ```

### 4. Stop the Mock Server
```bash
gockapi --config-path ./my-configs stop serviceA
```
- This will terminate the running mock server for `serviceA`.

---

## Running Multiple Services Independently

You can run multiple mock services at the same time, each in its own process. This allows you to start, stop, and interact with each service independently.

### 1. Create Multiple Configurations
For example, add `serviceB.json` and `serviceC.json`:

```json
{
  "service_name": "serviceB",
  "port": 55002,
  "endpoints": {
    "GET /api/bye": {
      "status_code": 200,
      "headers": {
        "Content-Type": "application/json"
      },
      "body": {
        "message": "Goodbye from serviceB!"
      }
    }
  }
}
```

```json
{
  "service_name": "serviceC",
  "port": 55003,
  "endpoints": {
    "GET /api/c": {
      "status_code": 200,
      "headers": {
        "Content-Type": "application/json"
      },
      "body": {
        "message": "Hello from serviceC!"
      }
    }
  }
}
```

### 2. Start Each Service in Its Own Process
```bash
gockapi --config-path ./my-configs start serviceA &
gockapi --config-path ./my-configs start serviceB &
gockapi --config-path ./my-configs start serviceC &
```
- Each service will run in its own background process.
- You can start as many as you want, each on its own port.

### 3. Test Each Service Independently
```bash
curl http://localhost:55001/api/hello   # serviceA
curl http://localhost:55002/api/bye     # serviceB
curl http://localhost:55003/api/c       # serviceC
```

### 4. Stop a Specific Service Without Affecting Others
```bash
gockapi --config-path ./my-configs stop serviceA
# serviceB and serviceC will keep running
gockapi --config-path ./my-configs stop serviceB
gockapi --config-path ./my-configs stop serviceC
```

---

## Using `stop-all`: Statelessly Stop All Services

You can stop all running mock servers in your config directory with a single command:

```bash
gockapi --config-path ./my-configs stop-all
```

**How it works:**
- The `stop-all` command is stateless: it scans all `*.json` config files in the target directory.
- For each config, it reads the port and attempts to kill any process on that port started by the tool.
- This works for servers started independently, even if you started them in separate shells or scripts.
- You will see a log for each service indicating whether it was stopped or if no process was found.

---

## Using `start-all`: All Services in a Single Process

You can also start all services in a directory at once using the `start-all` command:

```bash
gockapi --config-path ./my-configs start-all
```

**Important:**
- When you use `start-all`, all mock servers are started within the same process and context.
- If you terminate or stop one, you will terminate all of them at once.
- This is useful for quick setups, but for independent control, use the approach described above (starting each service in its own process).

---

## More Usage Examples

### Start Multiple Services (if you want only some services in Single)
```bash
gockapi --config-path ./my-configs start serviceA serviceB
```

---

## Commands
- `start-all` — Start all mock servers in the config directory (single process)
- `start <service>` — Start a specific service by name (without `.json`)
- `stop-all` — Stop all running mock servers (stateless, scans all configs)
- `stop <service>` — Stop a specific service
- `reload <service>` — Reload configuration for a service
- `status` — Show status of all services

---

## License
MIT