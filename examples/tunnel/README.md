# Cloudflare Tunnel Examples

This directory contains examples for managing Cloudflare Tunnels (formerly Argo Tunnel) through Crossplane.

## Cloudflare Tunnel

The `tunnel.yaml` file demonstrates how to create a Cloudflare Tunnel for secure remote access to internal services.

### Key Features Demonstrated

- **Secure Remote Access**: Expose internal services through Cloudflare's edge
- **Zero Trust Integration**: Combine with Access applications for authenticated access
- **Metadata Tagging**: Organize tunnels with custom metadata
- **Configuration Management**: Control tunnel behavior through configuration sources

### Usage

```bash
kubectl apply -f tunnel.yaml
```

### Prerequisites

- `cloudflared` daemon installed and authenticated
- Tunnel secret generated using `cloudflared tunnel login` and `cloudflared tunnel token`
- Cloudflare account with Tunnel capability

### Setup Steps

1. **Install cloudflared**:
   ```bash
   # Download and install cloudflared
   wget https://github.com/cloudflare/cloudflared/releases/latest/download/cloudflared-linux-amd64
   sudo mv cloudflared-linux-amd64 /usr/local/bin/cloudflared
   sudo chmod +x /usr/local/bin/cloudflared
   ```

2. **Authenticate**:
   ```bash
   cloudflared tunnel login
   ```

3. **Create tunnel secret**:
   ```bash
   # Generate a tunnel secret (base64 encoded)
   cloudflared tunnel token <tunnel-name>
   ```

4. **Apply the Crossplane resource**:
   ```bash
   kubectl apply -f tunnel.yaml
   ```

### Customization

- Replace `your-account-id` with your Cloudflare account ID
- Update `tunnelSecret` with the base64-encoded secret from step 3
- Modify metadata tags for organization and filtering
- Adjust `configSrc` based on your configuration management approach

### Configuration Sources

- **local**: Configuration managed locally via cloudflared config file
- **cloudflare**: Configuration stored in Cloudflare dashboard
- **warp**: Configuration managed through WARP client

### Related Resources

- **DNS Records**: Create CNAME records pointing to tunnel domains
- **Access Applications**: Secure tunnel endpoints with authentication
- **Load Balancers**: Distribute traffic across multiple tunnel instances