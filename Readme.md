---

## Flags for gRPC Clipboard Server
You can use the following flags when running the server:

- `--port`: Port to run the gRPC server on (default: `50005`).
- `--insecure`: Use insecure gRPC connection (default: `true`).
- `--cert`: Path to the TLS certificate file (required if `--insecure` is `false`).
- `--key`: Path to the TLS key file (required if `--insecure` is `false`).

### Example Usage:
- Run with insecure connection:
  ```bash
  ./clip_server --port 50005 --insecure true
  ```
- Run with secure connection:
  ```bash
  ./clip_server --port 50005 --insecure false --cert /path/to/cert --key /path/to/key
  ```

---

## Flags for gRPC Clipboard Client
You can use the following flags when running the client:

- `--ip`: IP address of the gRPC server (default: `localhost:50005`).
- `--insecure`: Use insecure gRPC connection (default: `true`).
- `--tls`: Path to TLS certificate (required if `--insecure` is `false`).
- `--group`: Group name (optional).

### Example Usage:
- Connect with insecure connection:
  ```bash
  ./clip_client --ip localhost:50005 --insecure true
  ```
- Connect with secure connection:
  ```bash
  ./clip_client --ip localhost:50005 --insecure false --tls /path/to/tls
  ```

---

## Additional Information
If you encounter issues with clipboard functionality, ensure the `xsel` package is installed:
```bash
apt install xsel
```

---

## Tested on Linux and Windows

---

## Future Work
- Group clipboard entries by user-defined groups.
- Implement clipboard history.
- Implement a web interface for clipboard management.
- Add support for image and file clipboard entries.
- Add support for clipboard entry expiration.
- Implement clipboard entry search functionality.
