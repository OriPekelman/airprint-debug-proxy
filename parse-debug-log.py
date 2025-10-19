#!/usr/bin/env python3
"""
Parse AirPrint proxy debug logs and display them in a more readable format.
"""

import sys
import re
from datetime import datetime

def parse_log(filename):
    """Parse the debug log file and extract key information."""

    with open(filename, 'r', encoding='utf-8', errors='ignore') as f:
        content = f.read()

    # Split into request/response pairs
    sections = re.split(r'={80,}', content)

    request_count = 0
    print("=" * 80)
    print("AirPrint Proxy Debug Log Summary")
    print("=" * 80)
    print()

    for i, section in enumerate(sections):
        if not section.strip():
            continue

        # Check if this is a request section
        if "REQUEST at" in section:
            request_count += 1

            # Extract timestamp
            timestamp_match = re.search(r'REQUEST at ([\d\-: .]+)', section)
            timestamp = timestamp_match.group(1) if timestamp_match else "Unknown"

            # Extract client info
            client_match = re.search(r'Client IP: ([^\|]+) \| MAC: (.+)', section)
            if client_match:
                client_ip = client_match.group(1).strip()
                client_mac = client_match.group(2).strip()
            else:
                client_ip = "Unknown"
                client_mac = "Unknown"

            # Extract HTTP method and path
            method_match = re.search(r'(GET|POST|PUT|DELETE|OPTIONS)\s+([^\s]+)\s+HTTP', section)
            if method_match:
                method = method_match.group(1)
                path = method_match.group(2)
            else:
                method = "Unknown"
                path = "Unknown"

            # Extract Content-Type
            content_type_match = re.search(r'Content-Type:\s*([^\n]+)', section)
            content_type = content_type_match.group(1).strip() if content_type_match else "Unknown"

            # Extract Content-Length
            content_length_match = re.search(r'Content-Length:\s*(\d+)', section)
            content_length = content_length_match.group(1) if content_length_match else "0"

            print(f"Request #{request_count}")
            print(f"  Time:           {timestamp}")
            print(f"  Client:         {client_ip} (MAC: {client_mac})")
            print(f"  Method:         {method} {path}")
            print(f"  Content-Type:   {content_type}")
            print(f"  Content-Length: {content_length} bytes")

            # Try to identify IPP operation from body
            if "printer-uri" in section:
                print(f"  IPP Operation:  Get-Printer-Attributes")
            elif "Create-Job" in section:
                print(f"  IPP Operation:  Create-Job")
            elif "Send-Document" in section:
                print(f"  IPP Operation:  Send-Document")
            elif "Print-Job" in section:
                print(f"  IPP Operation:  Print-Job")

            # Look for response in next section
            if i + 1 < len(sections):
                next_section = sections[i + 1] if i + 1 < len(sections) else ""
                response_match = re.search(r'HTTP/[\d.]+ (\d+) ([^\n]+)', next_section)
                if response_match:
                    status_code = response_match.group(1)
                    status_text = response_match.group(2).strip()
                    print(f"  Response:       {status_code} {status_text}")

                    # Extract response content length
                    resp_length_match = re.search(r'Content-Length:\s*(\d+)', next_section)
                    if resp_length_match:
                        resp_length = resp_length_match.group(1)
                        print(f"  Response Size:  {resp_length} bytes")

            print()

    print("=" * 80)
    print(f"Total Requests: {request_count}")
    print("=" * 80)

def main():
    if len(sys.argv) != 2:
        print("Usage: python3 parse-debug-log.py <debug-log-file>")
        print()
        print("Example:")
        print("  python3 parse-debug-log.py airprint-debug-2025-10-19-19-55-56.log")
        sys.exit(1)

    filename = sys.argv[1]

    try:
        parse_log(filename)
    except FileNotFoundError:
        print(f"Error: File '{filename}' not found")
        sys.exit(1)
    except Exception as e:
        print(f"Error parsing log: {e}")
        sys.exit(1)

if __name__ == "__main__":
    main()
