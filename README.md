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
- **Graceful Shutdown** via standard OS signals (SIGINT, SIGTERM).
- **Lightweight**: minimal external dependencies, just Go’s built-in `net/http` and `httputil`.

---

## Usage

1. **Build or download** the compiled binary (see [Build](#build)).
2. **Run** the service, passing in required arguments:
   ```bash
   ./airprint-proxy -target=<printer-host-or-ip> [-port=631]
   ```
3. **Example**:
   ```bash
   # Proxy IPP requests from all interfaces at :631
   # to a printer located at 10.0.0.50
   ./airprint-proxy -target=10.0.0.50
   ```
4. **Check logs** on your console or wherever you direct stdout. You’ll see messages indicating incoming requests and proxy events.

---

## Required Arguments

| Argument      | Default | Description                                                          |
|---------------|---------|----------------------------------------------------------------------|
| `-target`     | *none*  | The IP or DNS name of the destination printer or service to proxy to. **Required**. |
| `-port`       | `631`   | The TCP port to listen on for incoming AirPrint/IPP requests.        |

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

## License

This project is released under the [MIT License](./LICENSE).

---

*Happy printing!*
