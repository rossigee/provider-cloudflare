# Cloudflare Access Examples

This directory contains examples for managing Cloudflare Access (Zero Trust) applications through Crossplane.

## Access Application

The `access-application.yaml` file demonstrates how to create a Cloudflare Access application with authentication policies for secure application access.

### Key Features Demonstrated

- **Zero Trust Security**: Configure authentication policies for applications
- **Session Management**: Control session duration and cookie attributes
- **CORS Configuration**: Set up cross-origin resource sharing headers
- **Identity Provider Integration**: Connect with external identity providers
- **Policy Precedence**: Define multiple access policies with precedence ordering

### Usage

```bash
kubectl apply -f access-application.yaml
```

### Prerequisites

- Cloudflare account with Zero Trust enabled
- Identity providers configured in Cloudflare Access
- Access policies created in Cloudflare dashboard

### Customization

- Replace `your-account-id` with your actual Cloudflare account ID
- Update `your-idp-id` with your identity provider IDs
- Modify domains, policies, and CORS settings according to your requirements
- Adjust session duration and cookie attributes for your security needs

### Related Resources

- **Access Policies**: Define authentication rules in Cloudflare dashboard
- **Identity Providers**: Configure SAML, OAuth, or other authentication methods
- **Service Tokens**: Create tokens for API access without user interaction