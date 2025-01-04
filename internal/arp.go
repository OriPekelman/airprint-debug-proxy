package internal

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
)

// getMACFromIP attempts to retrieve the MAC address for the given IP.
// This is a best-effort approach:
// - On Linux, we read /proc/net/arp
// - On other systems, we fallback to `arp -a` and parse the output.
func getMACFromIP(ip string) (string, error) {
    if runtime.GOOS == "linux" {
        return getMACFromArpFile(ip)
    }
    return getMACFromArpCmd(ip)
}

// getMACFromArpFile is a Linux-specific approach parsing /proc/net/arp
func getMACFromArpFile(ip string) (string, error) {
    f, err := os.Open("/proc/net/arp")
    if err != nil {
        return "", err
    }
    defer f.Close()

    scanner := bufio.NewScanner(f)
    // Skip header
    scanner.Scan()

    for scanner.Scan() {
        line := scanner.Text()
        fields := strings.Fields(line)
        if len(fields) >= 4 && fields[0] == ip {
            return fields[3], nil
        }
    }

    return "", fmt.Errorf("not found in /proc/net/arp")
}

// getMACFromArpCmd calls `arp -a` and tries to parse the result
func getMACFromArpCmd(ip string) (string, error) {
    cmd := exec.Command("arp", "-a")
    var out bytes.Buffer
    cmd.Stdout = &out

    if err := cmd.Run(); err != nil {
        return "", err
    }

    lines := strings.Split(out.String(), "\n")
    for _, line := range lines {
        if strings.Contains(line, ip) {
            // typical line example: ? (192.168.1.10) at 00:11:22:33:44:55 on ...
            parts := strings.Split(line, " ")
            for i, part := range parts {
                if part == "at" && i+1 < len(parts) {
                    mac := parts[i+1]
                    return mac, nil
                }
            }
        }
    }

    return "", fmt.Errorf("MAC not found for IP %s", ip)
}
