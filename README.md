# AirPrint Proxy

A simple Golang service that **listens for AirPrint (IPP) requests on a specified port** and **proxies them** to a specified printer or service, while logging the IP and (best-effort) MAC address of the request.

> **Note**: AirPrint uses Internet Printing Protocol (IPP), commonly on **TCP port 631**.

---

## Features

- **Reverse Proxy** to forward AirPrint (IPP) requests to a specified target.
- **Logging** of:
  - Incoming request IP
  - MAC address (where available)
  - Proxy success/failure
- **Debug Mode** to capture complete wire communication:
  - Full HTTP request/response dumps
  - Headers, body, and metadata
  - Timestamped logs for analysis
- **Graceful Shutdown** via standard OS signals (SIGINT, SIGTERM).
- **Lightweight**: minimal external dependencies, just Go's built-in `net/http` and `httputil`.

---

## Usage

1. **Build or download** the compiled binary (see [Build](#build)).
2. **Run** the service, passing in required arguments:
   ```bash
   ./airprint-proxy -target=<printer-host-or-ip> [-port=631] [-debug]
   ```
3. **Example**:
   ```bash
   # Proxy IPP requests from all interfaces at :631
   # to a printer located at 10.0.0.50
   ./airprint-proxy -target=10.0.0.50

   # Enable debug mode to capture full communication stream
   ./airprint-proxy -target=10.0.0.50 -debug
   ```
4. **Check logs** on your console or wherever you direct stdout. You'll see messages indicating incoming requests and proxy events.
5. **Debug output**: When `-debug` is enabled, a timestamped log file (e.g., `airprint-debug-2025-01-15-14-30-00.log`) will be created in the current directory containing the complete wire communication.

---

## Command-Line Arguments

| Argument      | Default | Description                                                          |
|---------------|---------|----------------------------------------------------------------------|
| `-target`     | *none*  | The IP or DNS name of the destination printer or service to proxy to. **Required**. |
| `-port`       | `631`   | The TCP port to listen on for incoming AirPrint/IPP requests.        |
| `-debug`      | `false` | Enable debug mode to log complete HTTP request/response data to a timestamped file. |

If `-target` is not provided, the application will exit and print usage information.

---

## Base Requirements

- **Go 1.23+** (if building from source).
- **Port availability**: The application must be able to bind to the port specified by `-port`.
- **ARP lookup**:
  - On **Linux**: The service checks `/proc/net/arp`.
  - On **macOS / Windows**: The service attempts to parse the output of `arp -a`.
  - If ARP resolution fails or is unavailable, the MAC address will be logged as `unknown`.

---

## Build

### 1. Clone the Repository

```bash
git clone https://github.com/your-org/airprint-proxy.git
cd airprint-proxy
```

### 2. Install Dependencies

```bash
go mod tidy
```

### 3. Build Locally

```bash
go build -o airprint-proxy ./cmd/airprint-proxy
```

This generates the `airprint-proxy` executable in your local directory.

---

## Running

```bash
./airprint-proxy -target=printer.local -port=631
```

- Listens on all interfaces at port `631`.
- Proxies to `printer.local` on port `80` (by default, because we construct the target with `http://`).

---

## Example Logs

```
[INFO] Starting AirPrint proxy on port 631 -> 10.0.0.50
[INFO] Received AirPrint request from IP: 192.168.1.100, MAC: 00:AA:BB:CC:DD:EE
[INFO] Proxied AirPrint request from 192.168.1.100 -> 10.0.0.50
```

---

## Graceful Shutdown

- The service listens for `SIGINT` and `SIGTERM`.
- When it receives either of these signals, it attempts a graceful shutdown within 5 seconds.

---

## Adding a Printer to Use the Proxy

To route print jobs through the proxy for debugging, you need to add a printer that points to the proxy instead of directly to the printer.

### Method 1: Command Line (macOS/Linux)

Use `lpadmin` to add a new printer:

```bash
# Example: Proxying a Canon G650 printer
# First, start the proxy on a custom port (e.g., 6310)
./airprint-proxy -target=c080F8E00000.local:631 -port=6310 -debug

# In another terminal, add the printer pointing to the proxy
sudo lpadmin -p Canon_G650_Proxy -v ipp://localhost:6310/ipp/print -E -m everywhere
```

**Parameters:**
- `-p Canon_G650_Proxy` - Name of the printer as it will appear in your system
- `-v ipp://localhost:6310/ipp/print` - URI pointing to your proxy
- `-E` - Enable the printer
- `-m everywhere` - Use driverless IPP Everywhere printing

### Method 2: System Preferences (macOS)

1. Open **System Settings** → **Printers & Scanners**
2. Click the **+** button to add a printer
3. Select the **IP** tab
4. Fill in:
   - **Address:** `localhost`
   - **Protocol:** IPP (Internet Printing Protocol)
   - **Queue:** `/ipp/print`
   - **Port:** `6310` (or whatever port you used)
   - **Name:** Canon G650 Proxy (or any name you prefer)
   - **Use:** Select the matching printer driver or "Generic PCL/PostScript Printer"
5. Click **Add**

### Print Through the Proxy

Once added, select your proxy printer from any application and print normally. All communication will be logged to the debug file!

---

## Debug Mode

Debug mode allows you to capture and analyze the complete communication between clients and the printer. This is useful for:

- Troubleshooting printing issues
- Understanding the IPP protocol
- Analyzing printer behavior
- Security auditing

### Enabling Debug Mode

Simply add the `-debug` flag when starting the proxy:

```bash
# Example with Brother printer
./airprint-proxy -target=BRWC8A3E8495F40.local:631 -port=6310 -debug

# Example with Canon printer over IPPS
./airprint-proxy -target=c080F8E00000.local:631 -port=6310 -debug
```

### Debug Log Format

The debug log file will contain:
- **Timestamped entries** for each request/response pair
- **Client information** (IP address and MAC address)
- **Complete HTTP headers** (request and response)
- **Full request/response bodies** (including binary IPP data)
- **Separators** for easy reading

Example debug log excerpt:
```
================================================================================
REQUEST at 2025-01-15 14:30:45.123
Client IP: 192.168.1.100 | MAC: 00:AA:BB:CC:DD:EE
================================================================================
POST /ipp/print HTTP/1.1
Host: 10.0.0.50
Content-Type: application/ipp
Content-Length: 245
...

RESPONSE at 2025-01-15 14:30:45.234
--------------------------------------------------------------------------------
HTTP/1.1 200 OK
Content-Type: application/ipp
Content-Length: 128
...
```

### Debug Log Location

Debug logs are created in the current working directory with the filename format:
```
airprint-debug-YYYY-MM-DD-HH-MM-SS.log
```

For example: `airprint-debug-2025-01-15-14-30-00.log`

### Analyzing Debug Logs

Several tools can help you analyze the debug logs:

#### 1. Quick Summary with Python Parser

Use the included parser script for a high-level overview:

```bash
python3 parse-debug-log.py airprint-debug-2025-10-19-19-55-56.log
```

This shows a summary of all requests with timestamps, client IPs, and response codes.

#### 2. Terminal Viewing with less

Navigate the raw log file easily:

```bash
less airprint-debug-2025-10-19-19-55-56.log

# Search for specific requests:
# /REQUEST - Find next REQUEST
# n - Next match
# N - Previous match
# q - Quit
```

#### 3. Extract Specific Data with grep

Filter for specific information:

```bash
# See all HTTP status codes
grep "^HTTP/" airprint-debug-*.log

# Find all requests with timestamps
grep "^REQUEST at" airprint-debug-*.log

# Search for specific IPP operations
grep -i "print-job" airprint-debug-*.log
```

#### 4. Visual Analysis in Text Editor

Open in VS Code or your favorite editor for syntax highlighting and search:

```bash
code airprint-debug-2025-10-19-19-55-56.log
```

#### 5. IPP Protocol Analysis

Use CUPS' `ipptool` to validate IPP messages:

```bash
# Test the printer/proxy directly
ipptool -tv ipp://localhost:6310/ipp/print get-printer-attributes.test
```

---

## License

This project is released under the [MIT License](./LICENSE).

---

*Happy printing!*
